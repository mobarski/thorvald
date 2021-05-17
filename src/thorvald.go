// Copyright (c) 2021 Maciej Obarski
// Licensed under GNU General Public License v3.0

package main
  
import (
    "bufio"
    "fmt"
    //"log"
    "os"
	"time"
	"strings"
	"hash/crc32"
	"sort"
	"sync"
	"math"
	"flag"
	"io"
	//"runtime"
	"sync/atomic" // atomic.AddUint32(v, 1) <- (v *uint32)
)

func min(x, y int) int {
    if x < y {
        return x
    }
    return y
}

func max_u32(x, y uint32) uint32 {
    if x < y {
        return y
    }
    return x
}

// --- PROGRESS ---

type progress struct {
	total uint64
	label string
	unit  string
	done  uint64
	t0    time.Time
}

func Progress(total int, label string, unit string) progress {
	return progress{uint64(total), label, unit, 0, time.Now()}
}

func (p *progress) Add(x int) {
	atomic.AddUint64(&p.done,uint64(x))
	elapsed := time.Since(p.t0)
	rate := float64(p.done) / elapsed.Seconds()
	if p.total>0 {
		done_pct := float64(p.done) / float64(p.total) * 100
		//fmt.Fprintf(os.Stderr, "\r%s: %d / %d %s (%.f%%) -> %.1fs (%.1f %s/s)", p.label, p.done, p.total, p.unit, done_pct, elapsed.Seconds(), rate, p.unit)
		fmt.Fprintf(os.Stderr, "\r%s: %d %s -> %.1fs (%.1f %s/s) -> %.f%% done ", p.label, p.done, p.unit, elapsed.Seconds(), rate, p.unit, done_pct)
	} else {
		fmt.Fprintf(os.Stderr, "\r%s: %d %s -> %.1fs (%.1f %s/s) ", p.label, p.done, p.unit, elapsed.Seconds(), rate, p.unit)
	}
	// TODO: remaining
	// TODO: rate
}

func (p *progress) Close() {
	fmt.Fprintf(os.Stderr, "\n")
}

// --- SKETCH UTILS ---

func estimate_count(sketch_len int, v_max uint32) int {
	if sketch_len>0 {
		return int(math.Round(float64(sketch_len)/float64(v_max) * float64(4294967295)))
	} else {
		return 0
	}
}

func estimate_intersection(sketch_len int, sketch_cap int, a_cnt int, b_cnt int, v_max uint32) int {
	if a_cnt<sketch_cap && b_cnt<sketch_cap {
		return sketch_len
	} else {
		c := estimate_count(sketch_len, v_max)
		return min(c,min(a_cnt,b_cnt))
	}
}

func press_enter(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	var x string; fmt.Scanf("%s",&x)
}


func other_triangle_format(fmt []string) []string {
	out := make([]string,len(fmt))
	for i,x := range fmt {
		switch x {
			case "ida" : out[i] = "idb"
			case "idb" : out[i] = "ida"
			case "ia"  : out[i] = "ib"
			case "ib"  : out[i] = "ia"
			case "a"   : out[i] = "b"
			case "b"   : out[i] = "a"
			default    : out[i] = fmt[i]
		}
	}
	return out
}


func check(err error) {
	if err != nil {
		panic(err)
	}
}

type Cfg struct {
	input_path  string
	output_path string
	buf_cap     int
	sketch_cap  int
	item_col    int
	features_col   int
	output_fmt  string
	workers     int
	c_min       int
	header_in   bool
	header_out  bool
	header_part bool
	diagonal    bool
	full        bool
}




type Engine struct {
	cfg              Cfg
	features_by_item map[string]map[uint32]bool
	range_by_item    map[string]int
	all_features     map[uint32]bool
	feature_freq     map[uint32]int
	feature_idf      map[uint32]float64
	item_idf_sum     []float64
	item_idf_sqr     []float64
	buf              []byte
	items            []string
	output_fmt       []string
	other_fmt        []string
	use_idf          bool
	items_cnt        int
	all_features_cnt int
}


func (e *Engine) load() {
	// --- SET / KMV SKETCH CONSTRUCTION --------------------------------------
	
	crcTable := crc32.MakeTable(crc32.Castagnoli)
	file, err := os.Open(e.cfg.input_path)
	check(err)
	
	buf_cap := e.cfg.buf_cap*1024*1024
	e.buf = make([]byte, buf_cap)
	scanner := bufio.NewScanner(file)
	scanner.Buffer(e.buf, buf_cap)
	scanner.Split(bufio.ScanLines)
	if e.cfg.header_in {
		scanner.Scan()
	}
	
	// XXX
	// fo,err := os.Create("xxx.sketch.out.tsv")
	// check(err)
	// w := bufio.NewWriter(fo)
	
	pg := Progress(0,"LOAD","items")
	for scanner.Scan() {
		text := scanner.Text()
		rec := strings.Split(text, "\t")
		//fmt.Printf("len rec[0]=%d rec[1]=%d \n",len(rec[0]),len(rec[1]))
		item := rec[e.cfg.item_col-1]
		features := strings.Split(rec[e.cfg.features_col-1], ",") // TODO: check error
		
		// SKETCH
		features_hash := make([]int, 0, len(features)) // TODO: rename features_hash
		for _,feature := range features {
			hash := int(crc32.Checksum([]byte(feature), crcTable))
			features_hash = append(features_hash, hash)
			e.all_features[uint32(hash)] = true // not a sketch part -> for metrics
			e.feature_freq[uint32(hash)] += 1 // not a sketch part -> for inverse feature frequency
		}
		var sketch []int // TODO: rename sketch
		if e.cfg.sketch_cap>0 {
			sort.Ints(features_hash)
			sketch_len := min(len(features_hash), e.cfg.sketch_cap)
			sketch = features_hash[:sketch_len]
			// TODO: save sketch
			// XXX
			// fmt.Fprintf(w,"%s\t",item)
			// for _,x := range sketch[:len(sketch)-1] {
				// fmt.Fprintf(w,"%d,",x)
			// }
			// fmt.Fprintf(w,"%d\n",sketch[len(sketch)-1])
		} else {
			sketch = features_hash
		}
		
		// SET
		features_set := make(map[uint32]bool)
		for _,hash := range sketch {
			features_set[uint32(hash)] = true
		}
		e.features_by_item[item] = features_set
		e.range_by_item[item] = len(features_hash)
			
		pg.Add(1)
	}
	// w.Flush()
	// fo.Close()
	file.Close()
	pg.Close()
	
	err = scanner.Err()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: reading input failed (%v)\n", err)
		os.Exit(1)
	}

	//fmt.Printf("items[1] features cnt %d\n",len(e.features_by_item["f1"])) // XXX 
	//fmt.Printf("items[2] features cnt %d\n",len(e.features_by_item["f2"])) // XXX

	e.items_cnt = len(e.features_by_item)
	//e.items_cnt := 100 // XXX
	e.items = make([]string, 0, e.items_cnt)
	for name := range e.features_by_item {
		e.items = append(e.items, name)
	}
	sort.Strings(e.items)
}


func (e *Engine) calc_idf() {
	e.use_idf = true // TODO: only when weighted metric in output_fmt -> 12% better performance of item-item similarity
	
	e.feature_idf  = make(map[uint32]float64, len(e.feature_freq))
	e.item_idf_sum = make([]float64,len(e.items))
	e.item_idf_sqr = make([]float64,len(e.items))
	
	if e.use_idf {
		for u,freq := range e.feature_freq {
			e.feature_idf[u] = math.Log(float64(len(e.items)) / float64(freq))
		}
		pg := Progress(len(e.items)," IDF","items")
		for i,item := range e.items {
			sum := 0.0
			sqr := 0.0
			for u := range e.features_by_item[item] {
				idf := e.feature_idf[u]
				sum += idf
				sqr += idf*idf
			}
			e.item_idf_sum[i] = sum
			e.item_idf_sqr[i] = sqr
			pg.Add(1)
		}
		pg.Close()
	}
}


func (e *Engine) item_item(i int, j int, w io.Writer, partition int) {
	mi := e.items[i]
	mj := e.items[j]
	mi_cnt := e.range_by_item[mi] // exact value - not from sketch
	mj_cnt := e.range_by_item[mj] // exact value - not from sketch
	// --- INTERSECTION ---
	common_cnt := 0
	common_sum := 0.0
	common_sqr := 0.0
	v_max := uint32(0)
	if i==j {
		common_cnt = mi_cnt
	} else {
		// set intersection - iter on smaller
		var a map[uint32]bool
		var b map[uint32]bool
		if mi_cnt < mj_cnt {
			a = e.features_by_item[mi]
			b = e.features_by_item[mj]
		} else {
			b = e.features_by_item[mi]
			a = e.features_by_item[mj]
		}
		for u := range a {
			_,ok := b[u]
			if ok {
				common_cnt += 1
				v_max = max_u32(u, v_max)
				if e.use_idf {
					idf := e.feature_idf[u]
					common_sum += idf
					common_sqr += idf*idf
				}
			}
		}
	}

	// SKETCH intersection estimation
	common_cnt_raw := common_cnt
	if e.cfg.sketch_cap>0 && i!=j {
		common_cnt = estimate_intersection(common_cnt, e.cfg.sketch_cap, mi_cnt, mj_cnt, v_max)
	}
					
	// --- METRICS ---
	a        := mi_cnt
	b        := mj_cnt
	c        := common_cnt
	cos      := float64(c) / math.Sqrt(float64(a*b))
	jaccard  := float64(c) / float64(a+b-c)
	dice     := float64(2*c) / float64(a+b)
	logdice  := 14.0 + math.Log2(dice)
	overlap  := float64(c) / float64(min(a,b))
	lift     := float64(c) / float64(a*b) * float64(e.all_features_cnt)
	pmi      := math.Log(lift)
	npmi     := pmi / -math.Log(float64(c) / float64(e.all_features_cnt))
	anpmi    := math.Abs(npmi)
	wdice    := 0.0
	wcos     := 0.0
	woverlap := 0.0
	wjaccard := 0.0
	wc       := 0.0
	if e.use_idf {
		wcos = common_sqr / math.Sqrt(e.item_idf_sqr[i]*e.item_idf_sqr[j])
		wdice = 2.0*common_sum / (e.item_idf_sum[i] + e.item_idf_sum[j])
		woverlap = common_sum / math.Min(e.item_idf_sum[i], e.item_idf_sum[j])
		wjaccard = common_sum / (e.item_idf_sum[i] + e.item_idf_sum[j] - common_sum)
		wc = common_sum
	}

	// --- OUTPUT ---
	if c < e.cfg.c_min {
		return
	}
	if j==i && !e.cfg.diagonal {
		return
	}
	// 
	ffmt := "%.4f"
	format_list := make([][]string,2)
	format_list[0] = e.output_fmt // first triangle
	format_list[1] = e.other_fmt  // second triangle (not empty only when "-full")
	for _,format := range format_list {
		// TODO: col -> inty zamiast stringow, przekodowanie na poczatku programu
		for k,col := range format {
			switch col {
				case "ida"       : fmt.Fprintf(w, "%s", mi)
				case "idb"       : fmt.Fprintf(w, "%s", mj)
				case "ia"        : fmt.Fprintf(w, "%d", i)
				case "ib"        : fmt.Fprintf(w, "%d", j)
				case "a"         : fmt.Fprintf(w, "%d", a)
				case "b"         : fmt.Fprintf(w, "%d", b)
				// symetric
				case "partition" : fmt.Fprintf(w, "%d", partition)
				case "wcos"      : fmt.Fprintf(w, ffmt, wcos)
				case "cos"       : fmt.Fprintf(w, ffmt, cos)
				case "c"         : fmt.Fprintf(w, "%d", c)
				case "craw"      : fmt.Fprintf(w, "%d", common_cnt_raw)
				case "jaccard"   : fmt.Fprintf(w, ffmt, jaccard)
				case "dice"      : fmt.Fprintf(w, ffmt, dice)
				case "overlap"   : fmt.Fprintf(w, ffmt, overlap)
				case "lift"      : fmt.Fprintf(w, ffmt, lift)
				case "pmi"       : fmt.Fprintf(w, ffmt, pmi)
				case "npmi"      : fmt.Fprintf(w, ffmt, npmi)
				case "anpmi"     : fmt.Fprintf(w, ffmt, anpmi)
				case "logdice"   : fmt.Fprintf(w, ffmt, logdice)
				case "wdice"     : fmt.Fprintf(w, ffmt, wdice)
				case "wjaccard"  : fmt.Fprintf(w, ffmt, wjaccard)
				case "woverlap"  : fmt.Fprintf(w, ffmt, woverlap)
				case "wc"        : fmt.Fprintf(w, ffmt, wc)
			}
			if k==len(e.output_fmt)-1 {
				fmt.Fprint(w,"\n")
			} else {
				fmt.Fprint(w,"\t")
			}
		}
	}
}


func (e *Engine) calc_similarity() {
	e.output_fmt = strings.Split(e.cfg.output_fmt, ",")
	e.all_features_cnt = len(e.all_features)
	
	pg := Progress(e.items_cnt,"CALC","items")
	var wg sync.WaitGroup
	f := func(i0,ii int) {
		filename := e.cfg.output_path
		if e.cfg.workers>=2 {
			filename += fmt.Sprintf(".p%d",i0+1)
		}
		fo := os.Stdout
		if len(e.cfg.output_path)>0 {
			fo2,err := os.Create(filename)
			check(err)
			fo = fo2
		}
		w := bufio.NewWriter(fo)
		// other triangle format string (swaped asymetrical columns)
		e.other_fmt = make([]string,0)
		if e.cfg.full {
			e.other_fmt = other_triangle_format(e.output_fmt)
		}
		// output header
		if e.cfg.header_part || e.cfg.header_out && i0==0 {
			header := strings.Join(e.output_fmt,"\t")
			fmt.Fprintf(w, "%s\n", header)
		}
		// item-item loop
		for i:=i0; i<e.items_cnt; i+=ii {
			j0 := i // will be 0 when output will be reduced to top X only
			for j:=j0; j<e.items_cnt; j++ {
				e.item_item(i,j,w,i0)
			}
			pg.Add(1) // progress // TODO: ilosc intersekcji ???
		}
		w.Flush()
		fo.Close()
		wg.Done()
	}
	
	W := e.cfg.workers
	wg.Add(W)
	for i:=0; i<W; i++ {
		go f(i,W)
	}
	wg.Wait()
	pg.Close()
}


func (e *Engine) main() {

	e.features_by_item = make(map[string]map[uint32]bool)
	e.range_by_item = make(map[string]int)
	e.all_features = make(map[uint32]bool)
	e.feature_freq = make(map[uint32]int)

	e.load()
	e.calc_idf()
	e.calc_similarity()
	
	// press_enter("\npress ENTER to reclaim memory")
}


func (cfg *Cfg) parse_args() {
	flag.StringVar(&cfg.input_path,  "i",   "", "input path")
	flag.StringVar(&cfg.output_path, "o",   "", "output path prefix (partitions will have .pX suffix)")
	flag.StringVar(&cfg.output_fmt,  "f",   "ida,idb,cos", "output format")
	
	flag.BoolVar(&cfg.header_in,   "ih", false, "input header")
	flag.BoolVar(&cfg.header_out,  "oh", false, "output header")
	flag.BoolVar(&cfg.header_part, "ph", false, "partition header")
	
	flag.BoolVar(&cfg.diagonal,  "diag", false, "include diagonal in output")
	flag.BoolVar(&cfg.full,      "full", false, "full output (including diagonal and lower triangle) (TODO)")
	
	flag.IntVar(&cfg.buf_cap,      "buf",  100, "line buffer capacity in MB")
	flag.IntVar(&cfg.item_col,     "coli",   1, "1-based column number of item name")
	flag.IntVar(&cfg.features_col, "colf",   2, "1-based column number of features")
	flag.IntVar(&cfg.c_min,        "cmin",   1, "minimum number of common features to show in output")
	flag.IntVar(&cfg.workers,      "w",      1, "number of workers")
	flag.IntVar(&cfg.sketch_cap,   "k",      0, "KMV sketch capacity, zero for no KMV usage")
	
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of this program:\n")
		fmt.Fprintf(os.Stderr, "./thorvald -i input.tsv -o output.tsv\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	
	// --- cfg validation ---
	if cfg.output_path=="" && cfg.workers>1 {
		fmt.Fprintf(os.Stderr, "ERROR: cannot output to stdout when workers>1\n")
		os.Exit(1)
	}
	
	//cfg.rest = flag.Args()
	n_args := len(os.Args[1:])
	if n_args==0 {
		flag.Usage()
		os.Exit(1)
	}	
}


func main() {
	cfg := Cfg{}
	cfg.parse_args()
	
	e := Engine{}
	e.cfg = cfg
	e.main()
}
