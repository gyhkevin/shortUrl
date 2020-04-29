package main

func main() {
	a := App{}
	a.Initialize(GetEnv())
	a.Run(":8000")
}
