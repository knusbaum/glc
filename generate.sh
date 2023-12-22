#!/usr/bin/env bash

echo package glc

echo 'import "reflect"'

cat <<EOF
//go:noinline
func encend(cont func()) {
     cont()
}
EOF

cat <<EOF
//go:noinline
func encstart(id uint64, cont func()) {
     var bs [8]byte
     for i := range bs {
     	 b := byte(id & 0xFF)
	 id >>= 8
	 bs[i] = b
     }
     switch bs[0] {
EOF

for ii in {0..9} {a..f}; do
    for jj in {0..9} {a..f}; do
	cat <<EOF
    case 0x${ii}${jj}:
        enc${ii}${jj}(bs[1:], cont)
EOF
	
    done;
done;
echo "    }"
echo "}"

for i in {0..9} {a..f}; do
    for j in {0..9} {a..f}; do
	cat <<EOF
//go:noinline
func enc${i}${j}(vnext []byte, cont func()) {
    if len(vnext) == 0 {
       encend(cont)
       return
    }
    switch vnext[0] {
EOF

	for ii in {0..9} {a..f}; do
	    for jj in {0..9} {a..f}; do
		cat <<EOF
    case 0x${ii}${jj}:
        enc${ii}${jj}(vnext[1:], cont)
EOF
	
	    done;
	done;
	echo "    }"
	echo "}"
	    
    done;
done;

cat <<EOF

var encmap map[uintptr]byte
var encstartpc, encendpc uintptr
EOF

for ii in {0..9} {a..f}; do
    for jj in {0..9} {a..f}; do
	cat <<EOF
var enc${ii}${jj}pc = uintptr(reflect.ValueOf(enc${ii}${jj}).UnsafePointer())
EOF
    done;
done;

cat <<EOF
func valForPC(pc uintptr) (byte, bool) {
     switch pc {
EOF
for ii in {0..9} {a..f}; do
    for jj in {0..9} {a..f}; do
	cat <<EOF
    case enc${ii}${jj}pc:
        return 0x${ii}${jj}, true
EOF
    done;
done;

cat <<EOF
    default:
        return 0, false
    }
}
EOF


cat <<EOF
func init() {
     encmap = make(map[uintptr]byte)
     encstartpc = uintptr(reflect.ValueOf(encstart).UnsafePointer())
     encendpc = uintptr(reflect.ValueOf(encend).UnsafePointer())
EOF
for ii in {0..9} {a..f}; do
    for jj in {0..9} {a..f}; do
	cat <<EOF
	encmap[uintptr(reflect.ValueOf(enc${ii}${jj}).UnsafePointer())] = 0x${ii}${jj}
EOF

    done;
done;
 echo '}'
 
