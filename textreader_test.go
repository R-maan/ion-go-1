package ion

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"math/big"
	"testing"
	"time"
)

func TestIgnoreValues(t *testing.T) {
	r := NewTextReaderString("{skip: me, please: true}\n[skip, me, please]\nfoo")

	if !r.Next() {
		t.Fatal(r.Err())
	}
	if r.Type() != StructType {
		t.Fatalf("expected StructType, got %v", r.Type())
	}

	if !r.Next() {
		t.Fatal(r.Err())
	}
	if r.Type() != ListType {
		t.Fatalf("expected ListType, got %v", r.Type())
	}

	if !r.Next() {
		t.Fatal(r.Err())
	}
	if r.Type() != SymbolType {
		t.Fatalf("expected SymbolType, got %v", r.Type())
	}

	val, err := r.StringValue()
	if err != nil {
		t.Fatal(err)
	}
	if val != "foo" {
		t.Errorf("expected foo, got %v", val)
	}

	if r.Next() {
		t.Error("next returned true")
	}
}

func TestReadSexps(t *testing.T) {
	test := func(str string, f func(r Reader, t *testing.T)) {
		t.Run(str, func(t *testing.T) {
			r := NewTextReaderString(str)
			if !r.Next() {
				t.Fatal(r.Err())
			}
			if r.Type() != SexpType {
				t.Errorf("expected type=SexpType, got %v", r.Type())
			}

			if err := r.StepIn(); err != nil {
				t.Fatal(err)
			}

			f(r, t)

			if err := r.StepOut(); err != nil {
				t.Fatal(err)
			}

			if r.Next() {
				t.Errorf("next returned true")
			}
			if r.Err() != nil {
				t.Fatal(r.Err())
			}
		})
	}

	test("(\t)", func(r Reader, t *testing.T) {
		if r.Next() {
			t.Errorf("next returned true")
		}
		if r.Err() != nil {
			t.Fatal(r.Err())
		}
	})

	test("(foo)", func(r Reader, t *testing.T) {
		symbol(t, r, "foo")
	})

	test("(foo bar baz)", func(r Reader, t *testing.T) {
		symbol(t, r, "foo")
		symbol(t, r, "bar")
		symbol(t, r, "baz")
	})
}

func TestStructs(t *testing.T) {
	test := func(str string, f func(r Reader, t *testing.T)) {
		t.Run(str, func(t *testing.T) {
			r := NewTextReaderString(str)
			if !r.Next() {
				t.Fatal(r.Err())
			}
			if r.Type() != StructType {
				t.Errorf("expected type=StructType, got %v", r.Type())
			}

			if err := r.StepIn(); err != nil {
				t.Fatal(err)
			}

			f(r, t)

			if err := r.StepOut(); err != nil {
				t.Fatal(err)
			}

			if r.Next() {
				t.Errorf("next returned true")
			}
			if r.Err() != nil {
				t.Fatal(r.Err())
			}
		})
	}

	test("{\r\n}", func(r Reader, t *testing.T) {
		if r.Next() {
			t.Errorf("next returned true")
		}
		if r.Err() != nil {
			t.Fatal(r.Err())
		}
	})

	test("{foo: bar}", func(r Reader, t *testing.T) {
		symbol(t, r, "bar")
		if r.FieldName() != "foo" {
			t.Errorf("expected foo, got %v", r.FieldName())
		}
	})

	test("{foo: a, bar: b, baz: c}", func(r Reader, t *testing.T) {
		symbol(t, r, "a")
		if r.FieldName() != "foo" {
			t.Errorf("expected foo, got %v", r.FieldName())
		}

		symbol(t, r, "b")
		if r.FieldName() != "bar" {
			t.Errorf("expected bar, got %v", r.FieldName())
		}

		symbol(t, r, "c")
		if r.FieldName() != "baz" {
			t.Errorf("expected baz, got %v", r.FieldName())
		}
	})
}

func TestMultipleStructs(t *testing.T) {
	r := NewTextReaderString("{} {} {}")

	for i := 0; i < 3; i++ {
		if !r.Next() {
			t.Error("next returned false")
			t.Fatal(r.Err())
		}
		if r.Type() != StructType {
			t.Fatalf("expected struct, got %v", r.Type())
		}

		if err := r.StepIn(); err != nil {
			t.Fatal(err)
		}
		if r.Next() {
			t.Fatal("next returned true")
		}
		if err := r.StepOut(); err != nil {
			t.Fatal(err)
		}
	}

	if r.Next() {
		t.Fatal("next returned true")
	}
}

func TestNullStructs(t *testing.T) {
	r := NewTextReaderString("null.struct {}")

	if !r.Next() {
		t.Fatal(r.Err())
	}
	if !r.IsNull() {
		t.Error("expected null, got not-null")
	}

	if !r.Next() {
		t.Fatal(r.Err())
	}
	if r.IsNull() {
		t.Error("expected not-null, got null")
	}

	if r.Next() {
		t.Fatal("next returned true")
	}
}

func TestLists(t *testing.T) {
	test := func(str string, f func(r Reader, t *testing.T)) {
		t.Run(str, func(t *testing.T) {
			r := NewTextReaderString(str)
			if !r.Next() {
				t.Fatal(r.Err())
			}
			if r.Type() != ListType {
				t.Errorf("expected type=ListType, got %v", r.Type())
			}

			if err := r.StepIn(); err != nil {
				t.Fatal(err)
			}

			f(r, t)

			if err := r.StepOut(); err != nil {
				t.Fatal(err)
			}

			if r.Next() {
				t.Errorf("next returned true")
			}
			if r.Err() != nil {
				t.Fatal(r.Err())
			}
		})
	}

	test("[    ]", func(r Reader, t *testing.T) {
		if r.Next() {
			t.Fatal("next returned true")
		}
	})

	test("[foo]", func(r Reader, t *testing.T) {
		symbol(t, r, "foo")
		if r.Next() {
			t.Fatal("next returned true")
		}
	})

	test("[foo, bar, baz]", func(r Reader, t *testing.T) {
		symbol(t, r, "foo")
		symbol(t, r, "bar")
		symbol(t, r, "baz")
		if r.Next() {
			t.Fatal("next returned true")
		}
	})
}

func symbol(t *testing.T, r Reader, eval string) {
	next(t, r, SymbolType)

	val, err := r.StringValue()
	if err != nil {
		t.Fatal(err)
	}
	if val != eval {
		t.Errorf("expected %v, got %v", eval, val)
	}
}

func next(t *testing.T, r Reader, et Type) {
	if !r.Next() {
		t.Fatal(r.Err())
	}
	if r.Type() != et {
		t.Fatalf("expected %v, got %v", et, r.Type())
	}
}

func TestClobs(t *testing.T) {
	test := func(str string, eval []byte) {
		t.Run(str, func(t *testing.T) {
			r := NewTextReaderString(str)
			if !r.Next() {
				t.Error("next returned false")
				t.Fatal(r.Err())
			}
			if r.Type() != ClobType {
				t.Errorf("expected type=ClobType, got %v", r.Type())
			}

			val, err := r.ByteValue()
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(val, eval) {
				t.Errorf("expected %v, got %v", eval, val)
			}

			if r.Next() {
				t.Error("next returned true")
			}
			if r.Err() != nil {
				t.Error(r.Err())
			}
		})
	}

	test("{{\"\"}}", []byte{})
	test("{{ \"hello world\" }}", []byte("hello world"))
	test("{{'''hello world'''}}", []byte("hello world"))
	test("{{'''hello'''\n'''world'''}}", []byte("helloworld"))
}

func TestBlobs(t *testing.T) {
	test := func(str string, eval []byte) {
		t.Run(str, func(t *testing.T) {
			r := NewTextReaderString(str)
			if !r.Next() {
				t.Error("next returned false")
				t.Fatal(r.Err())
			}
			if r.Type() != BlobType {
				t.Errorf("expected type=BlobType, got %v", r.Type())
			}

			val, err := r.ByteValue()
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(val, eval) {
				t.Errorf("expected %v, got %v", eval, val)
			}

			if r.Next() {
				t.Error("next returned true")
			}
			if r.Err() != nil {
				t.Error(r.Err())
			}
		})
	}

	test("{{}}", []byte{})
	test("{{AA==}}", []byte{0})
	test("{{  SGVsbG8g\r\nV29ybGQ=  }}", []byte("Hello World"))
}

func TestTimestamps(t *testing.T) {
	test := func(str string, eval time.Time) {
		t.Run(str, func(t *testing.T) {
			r := NewTextReaderString(str)
			if !r.Next() {
				t.Error("next returned false")
				t.Fatal(r.Err())
			}
			if r.Type() != TimestampType {
				t.Errorf("expected type=TimestampType, got %v", r.Type())
			}

			val, err := r.TimeValue()
			if err != nil {
				t.Fatal(err)
			}
			if !val.Equal(eval) {
				t.Errorf("expected %v, got %v", eval, val)
			}

			if r.Next() {
				t.Error("next returned true")
			}
			if r.Err() != nil {
				t.Error(r.Err())
			}
		})
	}

	et := time.Date(2001, time.January, 1, 0, 0, 0, 0, time.UTC)
	test("2001T", et)
	test("2001-01T", et)
	test("2001-01-01", et)
	test("2001-01-01T", et)
	test("2001-01-01T00:00Z", et)
	test("2001-01-01T00:00:00Z", et)
	test("2001-01-01T00:00:00.000Z", et)
	test("2001-01-01T00:00:00.000+00:00", et)
}

func TestDoubles(t *testing.T) {
	test := func(str string, eval string) {
		t.Run(str, func(t *testing.T) {
			r := NewTextReaderString(str)
			if !r.Next() {
				t.Error("next returned false")
			}
			if r.Type() != DecimalType {
				t.Errorf("expected type=DecimalType, got %v", r.Type())
			}

			ee := MustParseDecimal(eval)

			val, err := r.DecimalValue()
			if err != nil {
				t.Fatal(err)
			}
			if !ee.Equal(val) {
				t.Errorf("expected %v, got %v", ee, val)
			}

			if r.Next() {
				t.Error("next returned true")
			}
			if r.Err() != nil {
				t.Error(r.Err())
			}
		})
	}

	test("123.", "123")
	test("123.0", "123")
	test("123.456", "123.456")
	test("123d2", "12300")
	test("123d+2", "12300")
	test("123d-2", "1.23")
}

func TestFloats(t *testing.T) {
	test := func(str string, eval float64) {
		t.Run(str, func(t *testing.T) {
			r := NewTextReaderString(str)
			if !r.Next() {
				t.Error("next returned false")
			}
			if r.Type() != FloatType {
				t.Errorf("expected type=FloatType, got %v", r.Type())
			}

			val, err := r.FloatValue()
			if err != nil {
				t.Error(err)
			}
			if val != eval {
				t.Errorf("expected %v, got %v", eval, val)
			}

			if r.Next() {
				t.Error("next returned true")
			}
			if r.Err() != nil {
				t.Error(r.Err())
			}
		})
	}

	test("1e100\n", 1e100)
	test("1.2e+0", 1.2)
	test("-123.456e-78", -123.456e-78)
	test("+inf", math.Inf(1))
	test("-inf", math.Inf(-1))
}

func TestInts(t *testing.T) {
	test := func(str string, m func(Reader) error) {
		t.Run(str, func(t *testing.T) {
			r := NewTextReaderString(str)
			if !r.Next() {
				t.Error("next returned false")
			}
			if r.Type() != IntType {
				t.Errorf("expected type=IntType, got %v", r.Type())
			}

			if err := m(r); err != nil {
				t.Error(err)
			}

			if r.Next() {
				t.Error("next returned true")
			}
			if r.Err() != nil {
				t.Error(r.Err())
			}
		})
	}

	test("null.int", func(r Reader) error {
		if !r.IsNull() {
			return errors.New("expected isnull=true, got false")
		}

		val, err := r.IntValue()
		if err != nil {
			return err
		}
		if val != 0 {
			return fmt.Errorf("expected 0, got %v", val)
		}

		return nil
	})

	testInt := func(str string, eval int) {
		test(str, func(r Reader) error {
			val, err := r.IntValue()
			if err != nil {
				return err
			}
			if val != eval {
				return fmt.Errorf("expected %v, got %v", eval, val)
			}
			return nil
		})
	}

	testInt("0", 0)
	testInt("12_345", 12345)
	testInt("-1_2_3_4_5", -12345)
	testInt("0b00_0101", 5)
	testInt("-0b00_0101", -5)
	testInt("0x01_02_0e_0F", 0x01020e0f)
	testInt("-0x0102_0e0F", -0x01020e0f)

	testInt64 := func(str string, eval int64) {
		test(str, func(r Reader) error {
			val, err := r.Int64Value()
			if err != nil {
				return err
			}
			if val != eval {
				return fmt.Errorf("expected %v, got %v", eval, val)
			}
			return nil
		})
	}

	testInt64("0x123_FFFF_FFFF", 0x123FFFFFFFF)
	testInt64("-0x123_FFFF_FFFF", -0x123FFFFFFFF)

	testBigInt := func(str string, estr string) {
		test(str, func(r Reader) error {
			val, err := r.BigIntValue()
			if err != nil {
				return err
			}

			eval, _ := (&big.Int{}).SetString(estr, 0)
			if eval.Cmp(val) != 0 {
				return fmt.Errorf("expected %v, got %v", eval, val)
			}

			return nil
		})
	}

	testBigInt("0xEFFF_FFFF_FFFF_FFFF", "0xEFFFFFFFFFFFFFFF")
	testBigInt("0xFFFF_FFFF_FFFF_FFFF", "0xFFFFFFFFFFFFFFFF")
	testBigInt("-0x1_FFFF_FFFF_FFFF_FFFF", "-0x1FFFFFFFFFFFFFFFF")
}

func TestStrings(t *testing.T) {
	r := NewTextReaderString(`foo::"bar" "baz" 'a'::'b'::'''beep''' '''boop''' null.string`)

	test := func(etas []string, eval string) {
		if !r.Next() {
			t.Fatal("next returned false")
		}

		if r.Type() != StringType {
			t.Fatalf("expected type=string, got type=%v", r.Type())
		}

		if !strequals(r.TypeAnnotations(), etas) {
			t.Errorf("expected tas=%v, got tas=%v", etas, r.TypeAnnotations())
		}

		val, err := r.StringValue()
		if err != nil {
			t.Fatal(err)
		}

		if val != eval {
			t.Errorf("expected val=%v, got val=%v", eval, val)
		}
	}

	test([]string{"foo"}, "bar")
	test(nil, "baz")
	test([]string{"a", "b"}, "beepboop")
	test(nil, "")

	if r.Next() {
		t.Errorf("next unexpectedly returned true")
	}
	if r.Err() != nil {
		t.Error(r.Err())
	}
}

func TestSymbols(t *testing.T) {
	r := NewTextReaderString("'null'::foo bar a::b::'baz' null.symbol")

	test := func(etas []string, eval string) {
		if !r.Next() {
			t.Fatal("next returned false")
		}

		if r.Type() != SymbolType {
			t.Fatalf("expected type=symbol, got type=%v", r.Type())
		}

		if !strequals(r.TypeAnnotations(), etas) {
			t.Errorf("expected tas=%v, got tas=%v", etas, r.TypeAnnotations())
		}

		val, err := r.StringValue()
		if err != nil {
			t.Fatal(err)
		}

		if val != eval {
			t.Errorf("expected val=%v, got val=%v", eval, val)
		}
	}

	test([]string{"null"}, "foo")
	test(nil, "bar")
	test([]string{"a", "b"}, "baz")
	test(nil, "")

	if r.Next() {
		t.Errorf("next unexpectedly returned true")
	}
	if r.Err() != nil {
		t.Error(r.Err())
	}
}

func strequals(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func TestSpecialSymbols(t *testing.T) {
	r := NewTextReaderString("null\nnull.struct\ntrue\nfalse\nnan")

	// null
	{
		if !r.Next() {
			t.Fatal("next returned false")
		}
		if r.Type() != NullType {
			t.Errorf("expected type=NullType, got %v", r.Type())
		}
		if !r.IsNull() {
			t.Error("expected isNull=true, got false")
		}
	}

	// null.struct
	{
		if !r.Next() {
			t.Fatal("next returned false")
		}
		if r.Type() != StructType {
			t.Errorf("expected type=StructType, got %v", r.Type())
		}
		if !r.IsNull() {
			t.Error("expected isNull=true, got false")
		}
	}

	// true
	{
		if !r.Next() {
			t.Fatal("next returned false")
		}
		if r.Type() != BoolType {
			t.Errorf("expected type=BoolType, got %v", r.Type())
		}
		val, err := r.BoolValue()
		if err != nil {
			t.Fatal(err)
		}
		if !val {
			t.Error("expected value=true, got false")
		}
	}

	// false
	{
		if !r.Next() {
			t.Fatal("next returned false")
		}
		if r.Type() != BoolType {
			t.Errorf("expected type=BoolType, got %v", r.Type())
		}
		val, err := r.BoolValue()
		if err != nil {
			t.Fatal(err)
		}
		if val {
			t.Error("expected value=false, got true")
		}
	}

	// nan
	{
		if !r.Next() {
			t.Fatal("next returned false")
		}
		if r.Type() != FloatType {
			t.Errorf("expected type=FloatType, got %v", r.Type())
		}
		val, err := r.FloatValue()
		if err != nil {
			t.Fatal(err)
		}
		if !math.IsNaN(val) {
			t.Errorf("expected value=NaN, got %v", val)
		}
	}

	if r.Next() {
		t.Error("next returned true")
	}
}
