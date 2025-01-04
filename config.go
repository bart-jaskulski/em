package main

type Config struct {
	GridColumns int
	GridRows    int
	MaxResults  int
}

var DefaultConfig = Config{
	GridColumns: 4,
	GridRows:    3,
	MaxResults:  12,
}
