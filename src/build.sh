GOOS=linux CGO_ENABLED=0 go build -a -ldflags '-w' xfertool.go
if [[ $? -ne 0 ]]; then
	echo "Failed To Compile"
	exit
fi

