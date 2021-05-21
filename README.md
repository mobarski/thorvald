# Thorvald
Similarity calculation engine for unary data (e.g. implicit feedback).
Designed for item-based collaborative filtering and simple content-based filtering (e.g. for cold-start prevention).
Named after Thorvald Sørensen.

**WARNING** - this is pre-alpha version.

# Intro

TODO

## Features

- multiple similarity metrics: cos, npmi, logdice, jaccard and more
- IDF-like weighting of features (both content-based and user-based)
- parallel processing
- KMV sketch-based acceleration
- ability to limit output to top N items
- easy deployment: single, statically-linked executable
- [unix philosophy](https://en.wikipedia.org/wiki/Unix_philosophy)


## Installation

```
git clone https://github.com/mobarski/thorvald
cd thorvald/src
go build thorvald.go
```


## Simple example

### Input file
```
item	users
i1	u1,u2,u3,u4,u5,u6,u7
i2	u1,u3,u5,u7,u9
i3	u2,u4,u6,u8
i4	u2,u3,u5
```

### Invocation
```thorvald -i input.tsv -o output.tsv -ih -oh```

### Output file
```
ida	idb	cos
i1	i2	0.6761
i1	i3	0.5669
i1	i4	0.6547
i2	i4	0.5164
i3	i4	0.2887
```

## Another example

### Input file
```
item	tags	users
i1	t1,t2	u1,u2,u3,u4,u5,u6,u7
i2	t1,t3	u1,u3,u5,u7,u9
i3	t2,t3	u2,u4,u6,u8
i4	t2,t4	u2,u3,u5
i5	t2,t3	
```

### Invocation (collaborative)
```thorvald -i input.tsv -o output.tsv -ih -oh -colf 3 -f ida,idb,wcos,lift```

### Output file
```
ida	idb	wcos	lift
i1	i2	0.4941	1.1429
i1	i3	0.4479	1.0714
i1	i4	0.4348	1.4286
i2	i4	0.2695	1.3333
i3	i4	0.1386	0.8333
```

### Invocation (content-based)
```thorvald -i input.tsv -o output.tsv -ih -oh -colf 2 -f ida,idb,wcos```

### Output file
```
ida	idb	wcos
i1	i2	0.8486
i1	i3	0.0947
i1	i4	0.0325
i1	i5	0.0947
i2	i3	0.4462
i2	i5	0.4462
i3	i4	0.0550
i3	i5	1.0000
i4	i5	0.0550
```

# CLI options

|       option | info                                                             |
| -----------: | ---------------------------------------------------------------- | 
|        **i** | input path                                                       |
|        **o** | output path (default: stdout)                                    |
|        **f** | output format, (default: ida,idb,cos)                            |
|        **w** | number of workers (default: 1)                                   |
|        **k** | KMV sketch capacity, 0 for not using sketches (default: 0)       |
|       **ih** | input has header                                                 |
|       **oh** | include header in output                                         |
|      **top** | output only top N results, 0 for all results (default: 0)        |
|   **topcol** | output column number for top N selection (1-based) (default: 3)  |
|      **buf** | line buffer capacity in MB (default: 100)                        |
|     **coli** | column number of item id (1-based) (default: 1)                  |
|     **colf** | column number of features (1-based) (default: 2)                 |
|     **cmin** | minimum number of common features to show in output (default: 1) |
|     **diag** | include diagonal in the output                                   |
|     **full** | include upper and lower triangle in the output                   |

## Output format

|        option | info                                                     |
| ------------: | -------------------------------------------------------- |
|       **ida** | id of item A                                             |
|       **idb** | id of item B                                             |
|        **ia** | index of item A                                          |
|        **ib** | index of item B                                          |
|         **a** | number of features of item A                             |
|         **b** | number of features of item B                             |
|         **c** | number of features common to item A and item B           |
|      **araw** | raw number of elements in sketch A (TODO)                |
|      **braw** | raw number of elements in sketch B (TODO)                |
|      **craw** | raw number of elements in intersection of sketch A and B |
|       **cos** | cosine similarity                                        |
|      **dice** | Sørensen–Dice index                                      |
|   **logdice** | logDice score                                            |
|   **jaccard** | Jaccard index                                            |
|   **overlap** | overlap                                                  |
|      **lift** | lift                                                     |
|     **ilift** | inverse lift                                             |
|       **pmi** | PMI - Pointwise Mutual Information                       |
|      **npmi** | NPMI - Normalized Pointwise Mutual Information           |
|        **wa** | IDF weighted features of A                               |
|        **wa** | IDF weighted features of B                               |
|        **wc** | IDF weighted common features of A and B                  |
|      **wcos** | IDF weighted cosine similarity                           |
|     **wdice** | IDF weighted Sørensen–Dice index                         |
|  **wjaccard** | IDF weighted Jaccard index                               |
|  **woverlap** | IDF weighted overlap                                     |
|     **wlift** | IDF weighted lift                                        |
|      **wpmi** | IDF weighted PMI                                         |
|     **wnpmi** | IDF weighted NPMI                                        |
|  **wlogdice** | IDF weighted logDice                                     |
| **partition** | partition/worker ID                                      |


# Performance

Time complexity of calculating item to item similarities: `time ~ I * I * FPI`
where `I` is the number of elements and `FPI` is the average number of features per item.

When using KMV sketches, this formula looks like this: `time ~ I * I * K`
where `K` is the capacity of the sketch.

For `I=10k` and `FPI=1k` similarity calculation time on 8 cores is around 3 minutes.
The time grows linearly with `FPI` and quadratic with `I`.

Selecting only top N result doubles the required time.


# Planed features

- item-item combinations reduction via item features
- sketch input / output / merge
- context
- distributed processing support

[//]: # (online .md editor: https://markdown-editor.github.io/ )
