package lab_info

import "testing"

func Test(t *testing.T){
	resExpected := "Лабораторна робота 1: Проектування та реалізація системи збірки"
	resTest := lab_info(1, "Проектування та реалізація системи збірки")
	if resTest != resExpected {
		t.Error("Invalid lab info!")
	}
}