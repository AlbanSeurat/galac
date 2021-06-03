galac
=======

The package provides Go bindings for ALAC decoder (Apple version)

### Usage 

```
go get github.com/AlbanSeurat/galac
```

### Demo 

Inspired from [alac-go](https://github.com/xlab/alac-go/alac) there is a sample program to test that 
ALAC are well played using via [portaudio-go](https://github.com/xlab/portaudio-go). 
You can download free samples m4a files on [hyperion-records](http://www.hyperion-records.co.uk/testfiles.asp).

#### To build

As there is some cpp file, you have to build first the static library for ALAC and then the go binary 

```
cd alac/codec
make 
cd - 
go build 
```

#### Usage
```
galac -f [filename.m4a]
```

### To Do 

* Automate the build of the library from go
* Remove output buffer as input in Decode function  
* Change portaudio binding to remove c-for-go dependency
* Improve main program to not leak and show a progress