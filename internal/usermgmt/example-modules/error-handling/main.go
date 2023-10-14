package main

/*func main() {

	test()
	return

	// line, column := 12, 24
	//
	// simpleSyntaxErr := errors.New("syntax error")
	// wrappedSyntaxErr := errors.Wrap(simpleSyntaxErr, fmt.Sprintf("%d:%d", line, column))
	// formattedSyntaxErr := fmt.Errorf("%d:%d: syntax error", line, column)
	//
	// fmt.Println(simpleSyntaxErr)
	// fmt.Println(wrappedSyntaxErr)
	// fmt.Println(formattedSyntaxErr)

	var syntaxErr error
	syntaxErr = SyntaxError{
		Line:   12,
		Column: 24,
	}
	fmt.Println(syntaxErr) // output: 12:24: syntax error

	return

	httpService := &http.UserService{}

	r := gin.Default()
	r.GET("/user", httpService.GetUser)
	r.PUT("/user", httpService.CreateUser)
}

func CheckSyntax(text string) error {
	return nil
}

func test() {
	err := CheckSyntax("test data")
	switch err := err.(type) {
	case nil:
		fmt.Println("no err, do next action")
		// Perform next action...
	case SyntaxError:
		fmt.Println("encounter a syntax error, we can get detail information from it to handle", err.Line, err.Column)
		// Perform next action...
	default:
		fmt.Println("encounter an error, but is a unexpected error type")
		// The more error types we handle, the more we can avoid entering this case
		// Perform next action....
	}
}

type SyntaxError struct {
	Line   int
	Column int
}

func (e SyntaxError) Error() string {
	return fmt.Sprintf("%d:%d: syntax error", e.Line, e.Column)
}*/
