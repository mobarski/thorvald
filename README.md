# Thorvald
Similarity calculation engine for item-based collaborative filtering on unary data.
Named after Thorvald Sørensen.

**WARNING** - this is pre-alpha version.

# Intro

TODO

## Example input file

```
item	users
i1	u1,u2,u3,u4,u5,u6
i2	u1,u3,u5,u7,u9
i3	u2,u4,u6,u8
i4	u1,u2,u3,u7,u8
```

## Invocation

```./thorvald -i input.tsv -o output.tsv -ih -oh```

## Example output file

```
aname	bname	c
i1	i1	6
i1	i2	3
i1	i3	3
i1	i4	3
i2	i2	5
i2	i4	3
i3	i3	4
i3	i4	2
i4	i4	5
```

## Performance

TODO

## Features

- multilpe similarity metrics
- parallel processing
- KMV sketch based acceleration

TODO

# CLI options

|       option | info |
| -----------: | ---- | 
|        **i** | input path |
|        **o** | output path prefix (TODO: partitions) |
|        **f** | output format, (default: aname,bname,c) |
|        **w** | number of workers (default: 1) |
|       **ih** | input has header |
|       **oh** | include header in output (TODO) |
|       **ph** | include header in each partition (TODO) |
|      **buf** | line buffer capacity in MB (default: 10) |
|     **coli** | column number of item name (1-based) (default: 1) |
|     **colu** | column number of users names (1-based) (default: 2) |
|     **cmin** | minimum number of common users to show in output (default: 1) |

## Output format

|        option | info |
| ------------: | ---- |
|     **aname** | name of item A |
|     **bname** | name of item B |
|        **ai** | index of item A |
|        **bi** | index of item B |
|        **ci** | result index in 1d array -> ai*num_items + bi |
| **partition** | partition/worker ID  |
|         **a** | number of users of item A |
|         **b** | number of users of item B |
|         **c** | number of users common to item A and item B |
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
|  **tanimoto** | Tanimoto index |

# Planed features

- sketch input
- sketch output
- sketch delta update
- distributed processing support
- better output limiter
- asymetrical metrics
- non-triangular output
- item,context,users input format
- item-item combinations reduction via item features

[//]: # (online .md editor: https://markdown-editor.github.io/)
