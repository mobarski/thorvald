cd src
go build thorvald.go
move thorvald.exe ..
cd ..
thorvald.exe -i C:\repo\tsv\data\v1-big.tsv -o C:\repo\tsv\output\v1-big.p%d.tsv -w 4 -k 1000
