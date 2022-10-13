package util

import (
	"os"
	"testing"

	"github.com/iamcathal/lunchMoneyN26/dtos"
	"gotest.tools/assert"
)

func TestMain(m *testing.M) {
	setupAppConfig()
	code := m.Run()
	os.Exit(code)
}

func TestDivMod(t *testing.T) {
	expectedBig := 1
	expectedRemainder := 4

	actualBig, actualRemainder := divmod(64, 60)

	assert.Equal(t, expectedBig, actualBig)
	assert.Equal(t, expectedRemainder, actualRemainder)
}

func TestGetMinutesAndSecondsLeftReturnsMinutesWhenItsMoreThanOne(t *testing.T) {
	expectedOutput := "1m 20s"

	actualOutput := getMinutesAndSecondsLeft(80)

	assert.Equal(t, expectedOutput, actualOutput)
}

func TestGetMinutesAndSecondsLefDoesntReturnsMinutesWhenItsLessThanOne(t *testing.T) {
	expectedOutput := "59s"

	actualOutput := getMinutesAndSecondsLeft(59)

	assert.Equal(t, expectedOutput, actualOutput)
}

func setupAppConfig() {
	appConfig := dtos.AppConfig{
		APIPassword: "eeeeee",
	}
	SetConfig(appConfig)
}
