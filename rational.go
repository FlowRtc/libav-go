
type Rational struct {
	inner C.AVRational
}

func AVRational(r unsafe.Pointer) Rational {
	rat := (*C.AVRational)(r)
	return Rational{
		*rat,
	}
}

func NewRational(num, den int) Rational {
	return Rational{
		C.AVRational{
			C.int(num), C.int(den),
		},
	}
}
func (r Rational) AVRational() C.AVRational {
	return r.inner
}

func (r Rational) SetNum(n int) {
	r.inner.num = C.int(n)
}

func (r Rational) SetDen(d int) {
	r.inner.den = C.int(d)
}
func (r Rational) Num() int {
	return int(r.inner.num)
}

func (r Rational) Den() int {
	return int(r.inner.den)
}
