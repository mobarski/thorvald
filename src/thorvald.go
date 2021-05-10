// Copyright (c) 2021 Maciej Obarski
// Licensed under GNU General Public License v3.0

// v4 -> parametry CLI
// v4 -> output format
// v4 -> item zamiast movie

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
		fmt.Printf("\r%s: %d / %d %s (%.f%%) -> %.1fs (%.1f %s/s)", p.label, p.done, p.total, p.unit, done_pct, elapsed.Seconds(), rate, p.unit)
	} else {
		fmt.Printf("\r%s: %d %s -> %.1fs (%.1f %s/s)", p.label, p.done, p.unit, elapsed.Seconds(), rate, p.unit)
	}
	// TODO: remaining
	// TODO: rate
}

func (p *progress) Close() {
	fmt.Printf("\n")
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
	fmt.Println(msg)
	var x string; fmt.Scanf("%s",&x)
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

func core(cfg *Cfg) {
	buf_cap      := cfg.buf_cap * 1024*1024
	sketch_cap   := cfg.sketch_cap
	input_path   := cfg.input_path
	output_path  := cfg.output_path
	item_col     := cfg.item_col-1
	features_col := cfg.features_col-1
	output_fmt   := strings.Split(cfg.output_fmt, ",")
	workers      := cfg.workers // TODO jezeli=0 to tyle ile cpu
	c_min        := cfg.c_min
	header_in    := cfg.header_in
	header_out   := cfg.header_out
	header_part  := cfg.header_part
	diagonal     := cfg.diagonal
	full         := cfg.full
	
	// --- SET / KMV SKETCH CONSTRUCTION --------------------------------------
	
	crcTable := crc32.MakeTable(crc32.Castagnoli)
	file, err := os.Open(input_path)
	check(err)
	
	buf := make([]byte, buf_cap)
	scanner := bufio.NewScanner(file)
	scanner.Buffer(buf, buf_cap)
	scanner.Split(bufio.ScanLines)
	if header_in {
		scanner.Scan()
	}
	
	features_by_item := make(map[string]map[uint32]bool)
	range_by_item := make(map[string]int)
	all_features := make(map[uint32]bool)
	
	//fo,err := os.Create("sketch.estimation.tsv") // TODO: co z tym plikiem ???
	//check(err)
	//w := bufio.NewWriter(fo)
	
	pg := Progress(0,"LOAD","items")
	for scanner.Scan() {
		text := scanner.Text()
		rec := strings.Split(text, "\t")
		//fmt.Printf("len rec[0]=%d rec[1]=%d \n",len(rec[0]),len(rec[1]))
		item := rec[item_col]
		features := strings.Split(rec[features_col], ",") // TODO: check error
		
		// SKETCH
		features_hash := make([]int, 0, len(features)) // TODO: rename features_hash
		for _,feature := range features {
			hash := int(crc32.Checksum([]byte(feature), crcTable))
			features_hash = append(features_hash, hash)
			all_features[uint32(hash)] = true // not a sketch part -> for metrics
		}
		var sketch []int // TODO: rename sketch
		if sketch_cap>0 {
			sort.Ints(features_hash)
			sketch_len := min(len(features_hash),sketch_cap)
			sketch = features_hash[:sketch_len]
		} else {
			sketch = features_hash
		}
		
		// SET
		features_set := make(map[uint32]bool)
		for _,hash := range sketch {
			features_set[uint32(hash)] = true
		}
		features_by_item[item] = features_set
		range_by_item[item] = len(features_hash)
		
		pg.Add(1)
	}
	//w.Flush()
	//fo.Close()
	file.Close()
	pg.Close()

	//fmt.Printf("items[1] features cnt %d\n",len(features_by_item["f1"])) // XXX 
	//fmt.Printf("items[2] features cnt %d\n",len(features_by_item["f2"])) // XXX
	
	
	// --- INTERSECTION -------------------------------------------------------
	
	items_cnt := len(features_by_item)
	//items_cnt := 100 // XXX
	items := make([]string, 0, items_cnt)
	for name := range features_by_item {
		items = append(items, name)
	}
	sort.Strings(items)
	
	all_features_cnt := len(all_features)
	pg = Progress(items_cnt,"CALC","items")
	var wg sync.WaitGroup
	f := func(i0,ii int) {
		filename := output_path
		if workers>=2 {
			filename += fmt.Sprintf(".p%d",i0+1)
		}
		fo,err := os.Create(filename)
		check(err)
		w := bufio.NewWriter(fo)
		//
		if header_part || header_out && i0==0 {
			header := strings.Join(output_fmt,"\t")
			fmt.Fprintf(w, "%s\n", header)
		}
		//
		for i:=i0; i<items_cnt; i+=ii {
			mi := items[i]
			mi_cnt := range_by_item[mi] // exact value - not from sketch
			j0 := i
			if full {
				j0 = 0
			}
			for j:=j0; j<items_cnt; j++ {
				
				// --- INTERSECTION ---
				common_cnt := 0
				mj := items[j]
				mj_cnt := range_by_item[mj] // exact value - not from sketch
				v_max := uint32(0)
				if i==j {
					common_cnt = mi_cnt
				} else {
					// set intersection - iter on smaller
					var a map[uint32]bool
					var b map[uint32]bool
					if mi_cnt < mj_cnt {
						a = features_by_item[mi]
						b = features_by_item[mj]
					} else {
						b = features_by_item[mi]
						a = features_by_item[mj]
					}
					for u := range a {
						_,ok := b[u]
						if ok {
							common_cnt += 1
							v_max = max_u32(u, v_max)
						}
					}
				}

				// SKETCH intersection estimation
				common_cnt_raw := common_cnt
				if sketch_cap>0 && i!=j {
					common_cnt = estimate_intersection(common_cnt, sketch_cap, mi_cnt, mj_cnt, v_max)
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
				lift     := float64(c) / float64(a*b) * float64(all_features_cnt)
				pmi      := math.Log(lift)
				npmi     := pmi / -math.Log(float64(c) / float64(all_features_cnt))
				anpmi    := math.Abs(npmi)

				// --- OUTPUT ---
				if c < c_min {
					continue
				}
				if j==i && !diagonal {
					continue
				}
				// TODO: limitowanie outputu na podstawie jakiejs metryki -> overlap? abs(npmi)?
				// TODO: col -> inty zamiast stringow, przekodowanie na poczatku programu
				for k,col := range output_fmt {
					switch col {
						case "aname"     : fmt.Fprintf(w, "%s", mi)
						case "bname"     : fmt.Fprintf(w, "%s", mj)
						case "ai"        : fmt.Fprintf(w, "%d", i)
						case "bi"        : fmt.Fprintf(w, "%d", j)
						case "ci"        : fmt.Fprintf(w, "%d", i*items_cnt+j)
						case "a"         : fmt.Fprintf(w, "%d", a)
						case "b"         : fmt.Fprintf(w, "%d", b)
						// symetric
						case "partition" : fmt.Fprintf(w, "%d", i0)
						case "c"         : fmt.Fprintf(w, "%d", c)
						case "craw"      : fmt.Fprintf(w, "%d", common_cnt_raw)
						case "cos"       : fmt.Fprintf(w, "%f", cos)
						case "jaccard"   : fmt.Fprintf(w, "%f", jaccard)
						case "dice"      : fmt.Fprintf(w, "%f", dice)
						case "overlap"   : fmt.Fprintf(w, "%f", overlap)
						case "lift"      : fmt.Fprintf(w, "%f", lift)
						case "pmi"       : fmt.Fprintf(w, "%f", pmi)
						case "npmi"      : fmt.Fprintf(w, "%f", npmi)
						case "anpmi"     : fmt.Fprintf(w, "%f", anpmi)
						case "logdice"   : fmt.Fprintf(w, "%f", logdice)
					}
					if k==len(output_fmt)-1 {
						fmt.Fprint(w,"\n")
					} else {
						fmt.Fprint(w,"\t")
					}
				}
			}
			pg.Add(1) // progress // TODO: ilosc intersekcji ???
		}
		w.Flush()
		fo.Close()
		wg.Done()
	}
	
	W := workers
	wg.Add(W)
	for i:=0; i<W; i++ {
		go f(i,W)
	}
	wg.Wait()
	pg.Close()
			
	// press_enter("\npress ENTER to reclaim memory")

	// for m := range features_by_item {
		// features_by_item[m] = nil
	// }
	// features_by_item = nil
	// runtime.GC()
	// debug.FreeOSMemory()

	//press_enter("\npress ENTER to finish") // XXX
	//fmt.Println(items[:10]) // XXX
	//fmt.Println(common[:10]) // XXX

}

func main() {
	cfg := Cfg{}
	
	flag.StringVar(&cfg.input_path,  "i",   "", "input path")
	flag.StringVar(&cfg.output_path, "o",   "", "output path template, %d will be replaced with partition number")
	flag.StringVar(&cfg.output_fmt,  "f",   "aname,bname,cos", "output format")
	
	flag.BoolVar(&cfg.header_in,   "ih", false, "input header")
	flag.BoolVar(&cfg.header_out,  "oh", false, "output header")
	flag.BoolVar(&cfg.header_part, "ph", false, "partition header")
	
	flag.BoolVar(&cfg.diagonal,  "diag", false, "include diagonal in output")
	flag.BoolVar(&cfg.full,      "full", false, "full output (including diagonal and lower triangle) (TODO)")
	
	flag.IntVar(&cfg.buf_cap,      "buf",   10, "line buffer capacity in MB")
	flag.IntVar(&cfg.item_col,     "coli",   1, "1-based column number of item name")
	flag.IntVar(&cfg.features_col, "colf",   2, "1-based column number of features")
	flag.IntVar(&cfg.c_min,        "cmin",   1, "minimum number of common features to show in output")
	flag.IntVar(&cfg.workers,      "w",      1, "number of workers")
	flag.IntVar(&cfg.sketch_cap,   "k",      0, "KMV sketch capacity, zero for no KMV usage")
	
	flag.Usage = func() {
		fmt.Printf("Usage of this program:\n")
		fmt.Printf("./thorvald -i input.tsv -o output.tsv\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	// TODO walidacja cfg
	core(&cfg)
}
