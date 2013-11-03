package rpn

import (
	"errors"
	"math"
	"strconv"
	"strings"
)

const (
	OP_STACK_SIZE  = 256
	NUM_STACK_SIZE = 32
)

var (
	stack_ptr int = -1
	stack     [OP_STACK_SIZE]byte
)

func push(op rune) {
	stack_ptr++
	stack[stack_ptr] = byte(op)
}

func pop() {
	stack_ptr--
}

func head() rune {
	return rune(stack[stack_ptr])
}

func stack_is_full() bool {
	return stack_ptr+1 >= OP_STACK_SIZE
}

func stack_not_empty() bool {
	return stack_ptr >= 0
}

func left_assoc(op rune) bool {
	return op == '+' || op == '-' || op == '*' || op == '/'
}

func right_assoc(op rune) bool {
	return op == '%' || op == '^'
}

func is_oper(op rune) bool {
	return op == '+' || op == '-' || op == '*' || op == '/' ||
		op == '^' || op == '%'
}

func is_digit(op rune) bool {
	return op >= '0' && op <= '9'
}

func priority(op rune) int {
	switch op {
	case '+', '-':
		return 1

	case '*', '/':
		return 2

	case '%', '^':
		return 3
	}

	return -1
}

func ToRPN(expression string) (rpnText string, wrongAt int, parseErr error) {
	var (
		pointCount           int
		prevRune             rune
		openParenthesisCount int
	)

	reader := strings.NewReader(expression)
	wrongAt = 0
	rpnText = ""

	for {
		r, runeSize, eol := reader.ReadRune()
		if eol != nil {
			break
		}

		switch r {
		case '(':
			if prevRune != 0 && prevRune != '(' && !is_oper(prevRune) {
				return "", wrongAt, errors.New(
					"Open parenthesis can't be after " +
						strconv.QuoteRune(prevRune))
			}
			if stack_is_full() {
				return "", wrongAt, errors.New("Complex expression")
			}
			push(r)
			prevRune = r
			openParenthesisCount++

		case ')':
			if prevRune != ')' && !is_digit(prevRune) {
				return "", wrongAt, errors.New(
					"Close parenthesis can't be after " +
						strconv.QuoteRune(prevRune))
			}
			if openParenthesisCount == 0 {
				return "", wrongAt, errors.New("Missing open parenthesis")
			}
			for head() != '(' {
				rpnText += string(head())
				pop()
			}
			pop()
			prevRune = r
			openParenthesisCount--
			pointCount = 0

		case '+', '-', '*', '/', '%', '^':
			if prevRune != ')' && !is_digit(prevRune) {
				return "", wrongAt, errors.New(
					"Operation can't be after " +
						strconv.QuoteRune(prevRune))
			}

			rpnText += " "

			for stack_not_empty() &&
				((left_assoc(r) && priority(head()) >= priority(r)) ||
					(right_assoc(r) && priority(head()) > priority(r))) {
				rpnText += string(head())
				pop()
			}
			if stack_is_full() {
				return "", wrongAt, errors.New("Complex expression")
			}
			push(r)
			prevRune = r
			pointCount = 0

		case '.':
			if prevRune == ')' {
				return "", wrongAt, errors.New("Number can't be after )")
			}
			if pointCount > 0 {
				return "", wrongAt, errors.New("Too much decimal points")
			}
			rpnText += string(r)
			prevRune = r
			pointCount++

		default:
			if is_digit(r) {
				if prevRune == ')' {
					return "", wrongAt, errors.New("Number can't be after )")
				}
				rpnText += string(r)
				prevRune = r
			} else {
				return "", wrongAt, errors.New("Wrong symbol: " +
					strconv.QuoteRune(r))
			}
		}

		wrongAt += runeSize
	}

	if openParenthesisCount > 0 {
		return "", wrongAt, errors.New("Missing close parenthesis")
	}

	for stack_not_empty() {
		rpnText += string(head())
		pop()
	}

	return rpnText, -1, nil
}

func get_num(reader *strings.Reader) float64 {
	var (
		s string
		f float64
	)

	for {
		r, _, eol := reader.ReadRune()
		if eol != nil {
			break
		}
		if r == '.' || is_digit(r) {
			s += string(r)
		} else {
			reader.UnreadRune()
			break
		}
	}

	f, _ = strconv.ParseFloat(s, 64)
	return f
}

func calculateRPN(rpnText string) (res float64, err error) {
	var (
		st_cnt int = 0
		st     [NUM_STACK_SIZE]float64
		f      float64
	)

	reader := strings.NewReader(rpnText)

	for {
		r, _, eol := reader.ReadRune()
		if eol != nil {
			break
		}

		switch {
		case is_oper(r):
			if st_cnt < 2 {
				return 0, errors.New("Can't do " +
					strconv.QuoteRune(r) + ", few operands")
			}
			st_cnt--

			switch r {
			case '+':
				st[st_cnt-1] = st[st_cnt-1] + st[st_cnt]

			case '-':
				st[st_cnt-1] = st[st_cnt-1] - st[st_cnt]

			case '*':
				st[st_cnt-1] = st[st_cnt-1] * st[st_cnt]

			case '/':
				st[st_cnt-1] = st[st_cnt-1] / st[st_cnt]

			case '^':
				st[st_cnt-1] = math.Pow(st[st_cnt-1], st[st_cnt])

			case '%':
				st[st_cnt-1] = math.Mod(st[st_cnt-1], st[st_cnt])
			}
		case r == '.' || is_digit(r):
			reader.UnreadRune()
			f = get_num(reader)
			if st_cnt >= NUM_STACK_SIZE {
				return 0, errors.New("Complex expression")
			}

			st[st_cnt] = f
			st_cnt++
		}
	}

	if st_cnt > 1 {
		return 0, errors.New("I need more operations")
	}

	return st[0], nil
}

func Calculate(expression string) (res float64, wrongAt int, parseErr error) {
	var s string

	s, wrongAt, parseErr = ToRPN(expression)
	if parseErr != nil {
		return 0, wrongAt, parseErr
	}

	res, _ = calculateRPN(s)

	return res, wrongAt, parseErr
}
