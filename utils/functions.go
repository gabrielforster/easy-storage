package utils

func Ternary[T any](condition bool, trueVal, falsyVal T) T {
  if condition {
    return trueVal
  }

  return falsyVal
}
