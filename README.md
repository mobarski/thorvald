# Thorvald
Similarity calculation engine for item-based collaborative filtering on unary data.
Named after Thorvald Sørensen.

**WARNING** - this is pre-alpha version.

# Intro

TODO

## Example input file

```
item	users
i1	u1,u2,u3,u4,u5,u6,u7
i2	u1,u3,u5,u7,u9
i3	u2,u4,u6,u8
i4	u2,u3,u5
```

## Invocation

```./thorvald -i input.tsv -o output.tsv -ih -oh```

## Example output file

```
aname	bname	cos
i1	i2	0.676123
i1	i3	0.566947
i1	i4	0.654654
i2	i4	0.516398
i3	i4	0.288675
```

## Features

- multiple similarity metrics: cos, npmi, logdice, jaccard and more
- parallel processing
- KMV sketch based acceleration
- easy deployment: single, statically-linked executable
- easy build process: `go build thorvald.go`
- [unix philosophy](https://en.wikipedia.org/wiki/Unix_philosophy)

# CLI options

|       option | info |
| -----------: | ---- | 
|        **i** | input path |
|        **o** | output path prefix (partitions will have .pX suffix) |
|        **f** | output format, (default: aname,bname,cos) |
|        **w** | number of workers (default: 1) |
|       **ih** | input has header |
|       **oh** | include header in output |
|       **ph** | include header in each partition |
|      **buf** | line buffer capacity in MB (default: 10) |
|     **coli** | column number of item name (1-based) (default: 1) |
|     **colf** | column number of features (1-based) (default: 2) |
|     **cmin** | minimum number of common features to show in output (default: 1) |
|     **diag** | include diagonal in the output |
|     **full** | include upper and lower triangle in the output |

## Output format

|        option | info |
| ------------: | ---- |
|     **aname** | name of item A |
|     **bname** | name of item B |
|        **ai** | index of item A |
|        **bi** | index of item B |
| **partition** | partition/worker ID  |
|         **a** | number of features of item A |
|         **b** | number of features of item B |
|         **c** | number of features common to item A and item B |
|      **araw** | raw number of elements in sketch A (TODO) |
|      **braw** | raw number of elements in sketch B (TODO) |
|      **craw** | raw number of elements in intersection of sketch A and B |
|       **cos** | cosine similarity |
|      **dice** | Sørensen–Dice index |
|   **logdice** | logDice score |
|   **jaccard** | Jaccard index |
|   **overlap** | overlap |
|      **lift** | lift |
|       **pmi** | PMI - Pointwise Mutual Information |
|      **npmi** | NPMI - Normalized Pointwise Mutual Information |
|     **anpmi** | absolute NPMI |

# Performance

TODO

# Planed features

- sketch input
- sketch output
- sketch delta update
- inverse feature frequency
- distributed processing support
- better output limiter
- item-item combinations reduction via item features
- item,context,users input format

[//]: # (online .md editor: https://markdown-editor.github.io/ )
