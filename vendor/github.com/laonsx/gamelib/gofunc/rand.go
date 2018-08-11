package gofunc

import (
	"math"
	"math/rand"
	"time"
)

func init() {

	rand.Seed(time.Now().UnixNano())
}

// RandInt 根据传入的参数随机一个数
// n: 大于0的整数
// 如果只有n参数 从1-n随机，v参数不为空从n-v[0]随机
func RandInt(n int, v ...int) (i int) {

	if len(v) != 0 {

		max := v[0]
		num := max - n + 1
		i = n + rand.Intn(num)
	} else if n > 0 {

		i = rand.Intn(n) + 1
	}

	return
}

/**
返回浮点随机数
*/
func Randfloat(min, max float64) float64 {

	randNo := rand.Float64()*(max-min) + min
	i, f := math.Modf(randNo * 100)
	if f >= 0.5 {

		i += 1
	}

	return i / 100
}

// RandIntKey 根据map随机出一个map key，map的value为权重
func RandIntKey(data map[int]int) int {

	var n int
	for _, v := range data {

		n += v
	}

	rn := RandInt(n)

	var cr int
	var k int

	for i, rate := range data {

		cr += rate
		if cr >= rn {

			k = i

			break
		}
	}

	return k
}

// RandStrKey 根据map随机出一个map key，map的value为权重
func RandStrKey(data map[string]int) string {

	var n int
	for _, v := range data {

		n += v
	}

	rn := RandInt(n)

	var cr int
	var k string

	for i, rate := range data {

		cr += rate
		if cr >= rn {

			k = i

			break
		}
	}

	return k
}

// ShuffleInt 打乱slice
func ShuffleInt(src []int) []int {

	dest := make([]int, len(src))
	perm := rand.Perm(len(src))

	for i, v := range perm {

		dest[v] = src[i]
	}

	return dest
}

// ShuffleStr 打乱slice
func ShuffleStr(src []string) []string {

	dest := make([]string, len(src))
	perm := rand.Perm(len(src))

	for i, v := range perm {

		dest[v] = src[i]
	}

	return dest
}
