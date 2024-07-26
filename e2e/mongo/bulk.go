package main

import (
	"context"
	"fmt"
)

func bulk(ctx context.Context) error {
	s, err := testSuite(ctx, true)
	if err != nil {
		return err
	}

	fmt.Println(s)

	return nil
}
