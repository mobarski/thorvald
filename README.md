# thorvald
Similarity calculation engine for item-based collaborative filtering on unary data.

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

## Invocation

```./thorvald -i input.tsv -o output-%d.tsv -wrk 2```

## Performance

TODO

## Features

TODO

- KMV sketch based acceleration
- parallel processing

# CLI options

|       option | info |
| -----------: | ---- | 
|        **i** | input path |
|        **o** | output path, %d will be replaced with partition number |
|      **fmt** | output format, (default: aname,bname,c) |
|      **buf** | line buffer capacity in MB (default: 10) |
|  **colitem** | column number of item name (1-based) (default: 1) |
| **colusers** | column number of users names (1-based) (default: 2) |
|      **wrk** | number of workers (default: 1) |
|     **cmin** | minimum number of common users to show in output (default: 1) |

## Output format

|        option | info |
| ------------: | ---- |
|         **a** | number of users of item A |
|        **ai** | index of item A |
|     **aname** | name of item A |
|     **anpmi** | absolute NPMI |
|         **b** | number of users of item B |
|        **bi** | index of item B |
|     **bname** | name of the |
|         **c** | number of users common to item A and item B |
|        **ci** | result index in 1d array -> ai*num_items + bi |
|       **cos** | cosine similarity |
|      **craw** | raw number of elements in intersection of sketch A and B |
|      **dice** | Soersen-Dice index |
|   **jaccard** | Jaccard index |
|      **lift** | lift |
|   **logdice** | logDice score |
|      **npmi** | NPMI - Normalized Pointwise Mutual Information |
|   **overlap** | overlap |
| **partition** | partition/worker ID  |
|       **pmi** | PMI - Pointwise Mutual Information |
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
