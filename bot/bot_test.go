package bot_test

import (
	"testing"

	"github.com/Bluskript/sauce-loader/bot"
	"github.com/stretchr/testify/assert"
)

func TestSave(t *testing.T) {
	b, _ := bot.New("", "", "./sauce")
	assert.NoError(t, b.Save("https://cdn.discordapp.com/attachments/871757339194163273/871846774514520064/83883964_p1.jpg"))
}
