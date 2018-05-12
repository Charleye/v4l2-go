# v4l2-go
Encapsulate userspace V4L2 API with golang

# a few Word of Caution
first generate arch-specific golang source code. e.g.
for amd64 arch 
```bash
gcc -o tools/arch-go tools/arch-go.c
./tools/arch-go > amd64.go 
```
At the same, need to execute the process for arm, arm64

# example


