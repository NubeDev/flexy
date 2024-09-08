package helpers

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"math/rand"
	"strconv"
	"strings"
	"time"
	"unicode"
)

func UUID(prefix ...string) string {
	u, err := uuid.NewUUID()
	if err != nil {
		return time.Now().Format(time.StampNano)
	}
	// Convert the UUID to a string and remove hyphens
	uuidString := strings.ReplaceAll(u.String(), "-", "")
	uuidString = shuffleCharacters(uuidString)
	shortUUID := uuidString[:16]
	return fmt.Sprintf("i%s", shortUUID)
}

func shuffleCharacters(word string) string {
	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	// Convert the word to a slice of characters
	characters := []rune(word)

	// Find the index of the first letter in the original word
	firstLetterIndex := 0
	for i, char := range characters {
		if unicode.IsLetter(char) {
			firstLetterIndex = i
			break
		}
	}

	// Randomly shuffle the characters starting from the first letter using Fisher-Yates algorithm
	for i := len(characters) - 1; i > firstLetterIndex; i-- {
		j := rand.Intn(i-firstLetterIndex+1) + firstLetterIndex
		characters[i], characters[j] = characters[j], characters[i]
	}

	// Convert the shuffled slice back to a string
	shuffledWord := string(characters)

	return shuffledWord
}

func ProcessID(input string) (int, error) {
	// Check if the first two letters start with "R-"
	if !strings.HasPrefix(input, "R-") {
		return 0, errors.New("string does not start with 'R-'")
	}

	// Remove "R-" prefix
	input = input[2:]

	// Try to parse the remaining string as an integer
	num, err := strconv.Atoi(input)
	if err != nil {
		return 0, err
	}

	return num, nil
}
