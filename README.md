# lastnlines

tail the last N lines of a watched file

```
func main() {
	ln := NewLNL("./test.txt")
	ln.SetDepth(10)
	ln.Start()
	iter := 0
	stop := false
	for {
		iter += 1
		fmt.Println(iter)
		for _, line := range ln.Get() {
			fmt.Println(line)
			if strings.Contains(line, "stop_watching") {
				ln.Stop()
				stop = true
			}
		}
		if stop {
			break
		}
		time.Sleep(3 * time.Second)
	}
}
```