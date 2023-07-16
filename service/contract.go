package main

type Database interface {
	Transaction() (interface{}, error)
	Session() (interface{}, error)
}

type Storage interface {
}
