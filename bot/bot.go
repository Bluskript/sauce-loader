package bot

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/rapidloop/skv"
)

type Bot struct {
	*discordgo.Session
	prefix        string
	targetChannel string
	targetFolder  string
	store         *skv.KVStore
}

type CommandHandler func(m *discordgo.MessageCreate) string

func New(token string, channel string, folder string, store *skv.KVStore) (*Bot, error) {
	s, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	b := &Bot{
		Session:       s,
		prefix:        "%",
		targetChannel: channel,
		targetFolder:  folder,
		store:         store,
	}

	b.Identify.Intents = discordgo.IntentsAll

	b.AddHandler(b.OnMessage)
	b.AddHandler(b.CommandHandler)

	return b, nil
}

func (b *Bot) save(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	rawHash := md5.Sum(content)
	sum := hex.EncodeToString(rawHash[:])

	extension := filepath.Ext(resp.Request.URL.Path)
	savePath := path.Join(b.targetFolder, sum+extension)
	out, err := os.Create(savePath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = out.Write(content)
	if err != nil {
		return err
	}

	return nil
}

func (b *Bot) Save(attachment string, wg *sync.WaitGroup) {
	defer wg.Done()
	if err := b.save(attachment); err != nil {
		fmt.Printf("Failed to save image %v: %v\r\n", attachment, err)
	} else {
		fmt.Println("Saved " + attachment)
	}
}

func (b *Bot) saveAll(m *discordgo.MessageCreate) string {
	b.ChannelMessageSend(m.ChannelID, "Saving all the sauce...")
	var lastSeenID string
	reachedLastSeen := false
	b.store.Get("last_seen", &lastSeenID)
	b.store.Put("last_seen", m.ID)
	targetID := m.ID
	var wg sync.WaitGroup
	for {
		if reachedLastSeen {
			break
		}

		msgs, err := b.ChannelMessages(b.targetChannel, 50, targetID, "", "")
		if err != nil {
			return fmt.Sprintf("Error loading messages: %v", err)
		}
		if len(msgs) < 50 {
			break
		}
		targetID = msgs[len(msgs)-1].ID
		for _, msg := range msgs {
			if msg.ID == lastSeenID {
				reachedLastSeen = true
			}
			for _, a := range msg.Attachments {
				wg.Add(1)
				go b.Save(a.URL, &wg)
			}
		}
	}
	wg.Wait()
	return "Done saving"
}

func (b *Bot) CommandHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	c := m.Message.Content
	if !strings.HasPrefix(c, b.prefix) {
		return
	}
	items := strings.Split(
		strings.TrimPrefix(c, b.prefix),
		" ",
	)

	cmd := items[0]

	cmdHandlers := map[string]CommandHandler{
		"saveall": b.saveAll,
	}

	handler := cmdHandlers[cmd]

	if handler == nil {
		b.ChannelMessageSend(m.ChannelID, "Unknown command")
	}

	b.ChannelMessageSend(m.ChannelID, handler(m))
}

func (b *Bot) OnMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.ChannelID != b.targetChannel {
		return
	}

	var wg sync.WaitGroup

	for _, a := range m.Attachments {
		wg.Add(1)
		go b.Save(a.URL, &wg)
	}

	wg.Wait()
}
