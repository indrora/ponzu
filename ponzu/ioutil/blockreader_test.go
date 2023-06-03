package ioutil_test

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/indrora/ponzu/ponzu/ioutil"
)

func TestBlockReader_ReadBlock(t *testing.T) {
	testCases := []struct {
		name      string
		data      []byte
		chunkSize uint64
		expected  [][]byte
		expectErr error
	}{
		{
			name:      "read data with small chunk size",
			data:      []byte("Hello, world! This is a test."),
			chunkSize: 5,
			expected: [][]byte{
				[]byte("Hello"),
				[]byte(", wor"),
				[]byte("ld! T"),
				[]byte("his i"),
				[]byte("s a t"),
				[]byte("est."),
			},
			expectErr: io.EOF,
		},
		{
			name:      "read exact multiple",
			data:      []byte("1234567890"),
			chunkSize: 5,
			expected: [][]byte{
				[]byte("12345"),
				[]byte("67890"),
			},
			expectErr: io.EOF,
		},
		{
			name:      "read data with large chunk size",
			data:      []byte("Hello, world! This is a test."),
			chunkSize: 100,
			expected: [][]byte{
				[]byte("Hello, world! This is a test."),
			},
			expectErr: io.EOF,
		},
		{
			name:      "empty data",
			data:      []byte{},
			chunkSize: 10,
			expected:  [][]byte{},
			expectErr: io.EOF,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			var err error

			r := bytes.NewReader(tc.data)
			br := ioutil.NewBlockReader(r, tc.chunkSize)

			blocks := make([][]byte, 0)

			for {
				block, oerr := br.ReadBlock()
				if block != nil {
					blocks = append(blocks, block)
				}

				err = oerr
				if len(tc.expected) == len(blocks) {
					if !errors.Is(err, tc.expectErr) {
						t.Fatalf("unexpected error; want %v, got %v", tc.expectErr, err)
					} else {
						break
					}
				} else if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

			}

			if len(tc.expected) != len(blocks) {
				t.Fatalf("unexpected number of blocks: got %d, want %d", len(blocks), len(tc.expected))
			}

			for i, expectedBlock := range tc.expected {
				check_block := blocks[i]
				if !bytes.Equal(expectedBlock, check_block) {
					t.Errorf("unexpected block at index %d: got %q, want %q", i, check_block, expectedBlock)
				}
			}

			if !errors.Is(err, tc.expectErr) {
				t.Errorf("unexpected error: got %v, want %v", err, tc.expectErr)
			}

		})
	}
}
