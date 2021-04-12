package lab_info

import (
	"fmt"
	"strconv"
)

func lab_info (n int, name string) string {
	result := "Лабораторна робота " + strconv.Itoa(n) + ": " + name
	return result
}

func main() {
	lab := lab_info(1, "Проектування та реалізація системи збірки")
	fmt.Println(lab)	
}
