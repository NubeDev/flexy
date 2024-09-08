package helpers

import (
	"fmt"
	"github.com/google/uuid"
	"math/rand"
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

	// Find the public of the first letter in the original word
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
