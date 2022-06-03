package main
// f24298bcb77a7b57d5a1b95d34d6066f93d5f5357dcfff0df463da6ac66d
// https://edit.telegra.ph/auth/ZV2fpNHAJVwWXi80jvkCMdlT7NKiNPYLKY07JFFsyb
import (
	"log"
	"io"
	"net/http"
	"feed_reader/rss"
	"fmt"
	"encoding/json"
	"strconv"
	"strings"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	InitialPage = 0
	EntriesPerPage = 3
)

var validCommands = map[string]bool{
	"add": true,
	"list": true,
}

type SelectMenu interface {
	Len() int
	Offset() int
	Page() int
	TextOf(int) string
	CallbackDataOf(int) string
	PageInc()
	PageDec()
}

type SelectMenuBase struct {
	page int
}

func (b SelectMenuBase) Page() int {
	return b.page
}

func (b SelectMenuBase) Offset() int {
	return b.page * EntriesPerPage
}

func (b *SelectMenuBase) PageInc() {
	b.page++
}

func (b *SelectMenuBase) PageDec() {
	b.page--
}

type FeedSelectMenu struct {
	*SelectMenuBase
	feeds []*rss.Document
}

func (fsm FeedSelectMenu) Len() int {
	return len(fsm.feeds)
}

func (fsm FeedSelectMenu) TextOf(index int) string {
	return fsm.feeds[index].Channel.Title
}

func (fsm FeedSelectMenu) CallbackDataOf(index int) string {
	return "feed:" + fmt.Sprintf("%d", index)
}

type ItemSelectMenu struct {
	*SelectMenuBase
	// @todo: Try chaning to a slice of pointers
	items []rss.Item
}

func (ism ItemSelectMenu) Len() int {
	return len(ism.items)
}

func (ism ItemSelectMenu) TextOf(index int) string {
	if ism.items[index].Title == "" {
		return ism.items[index].Description.Value[:10] + "..."
	}

	return ism.items[index].Title
}

func (ism ItemSelectMenu) CallbackDataOf(index int) string {
	return "item:" + fmt.Sprintf("%d", index)
}

func (fsm *FeedSelectMenu) AddFeed(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		log.Panic(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	d, err := rss.NewDocument(body)
	if err != nil {
		log.Panic(err)
	}

	if d.SelfLink() == "" {
		d.SetSelfLink(url)
	}

	json, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		log.Panic(err)
	}
	fmt.Println(string(json))

	fsm.feeds = append(fsm.feeds, d)

	return nil
}

func NextPage(sm SelectMenu) {
	if sm.Offset() + EntriesPerPage < sm.Len() {
		sm.PageInc()
	}
}

func PrevPage(sm SelectMenu) {
	if sm.Page() > 0 {
		sm.PageDec()
	}
}

func PageMarkup(sm SelectMenu) tg.InlineKeyboardMarkup {
	var markup [][]tg.InlineKeyboardButton

	offset := sm.Offset()
	for i := offset; i < sm.Len() && i < offset + EntriesPerPage; i++ {
		entry := []tg.InlineKeyboardButton{
			tg.NewInlineKeyboardButtonData(
				sm.TextOf(i), sm.CallbackDataOf(i),
			),
		}
		markup = append(markup, entry)
	}

	var nav []tg.InlineKeyboardButton
	nav = append(nav, tg.NewInlineKeyboardButtonData("️⬆️", "back"))
	nav = append(nav, tg.NewInlineKeyboardButtonData("⬅️", "previous"))
	nav = append(nav, tg.NewInlineKeyboardButtonData(
		fmt.Sprintf("%d", sm.Page() + 1), "X"),
	)

	if sm.Offset() + EntriesPerPage < sm.Len() {
		nav = append(nav, tg.NewInlineKeyboardButtonData("️➡️", "next"))
	}

	markup = append(markup, nav)
	return tg.NewInlineKeyboardMarkup(markup...)
}

type NodeElement struct {
	tag string
	isTextNode bool
	attrs map[string]string
	children []NodeElement
}

var telegrapHTMLTags = map[string]struct{}{
	"a": struct{}{},
	"aside": struct{}{},
	"b": struct{}{},
	"blockquote": struct{}{},
	"br": struct{}{}, // SelfClosingTagToken
	"code": struct{}{},
	"em": struct{}{},
	"figcaption": struct{}{},
	"figure": struct{}{},
	"h3": struct{}{},
	"h4": struct{}{},
	"hr": struct{}{}, // SelfClosingTagToken
	"i": struct{}{},
	"iframe": struct{}{}, // @todo: Test which things you can embed?
	"img": struct{}{}, //@todo: Is it a SelfClosingTagToken element?
	"li": struct{}{},
	"ol": struct{}{},
	"p": struct{}{},
	"pre": struct{}{},
	"s": struct{}{},
	"strong": struct{}{},
	"u": struct{}{},
	"ul": struct{}{},
	"video": struct{}{},
}

var telegrapHTMLAttrs = map[string]struct{}{
	"href": struct{}{},
	"src": struct{}{},
}

func toTelegraphNodeElement(string html) {
	z := html.NewTokenizer(html)	

	var ne NodeElement

	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			return z.Err()
		case html.TextToken:
			fmt.Println(tn)
		case html.StartTagToken, html.EndTagToken, html.SelfClosingTagToken:
			tn, _ := z.TagName()
			if _, ok := telegrapHTMLTags[tn]; ok {
				fmt.Println(tn)
			}
		}
	}
}

func main() {
	var (
		itemSelectMenu ItemSelectMenu
		feedSelectMenu FeedSelectMenu
		selectMenu SelectMenu
	)

	itemSelectMenu.SelectMenuBase = &SelectMenuBase{page: 0}
	feedSelectMenu.SelectMenuBase = &SelectMenuBase{page: 0}

	bot, err := tg.NewBotAPI("1998517012:AAFb6iHQvtslbqIDgpsiJMSIrQ3PHrbq1LQ")
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tg.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		var c tg.Chattable

		switch {
		case update.Message != nil && validCommands[update.Message.Command()]:
			var err error
			cnf := tg.NewMessage(update.Message.Chat.ID, "")

			cnf.ReplyToMessageID = update.Message.MessageID

			if update.Message.Command() == "add" {
				cnf.Text = "has been added..."
				err = feedSelectMenu.AddFeed(update.Message.CommandArguments())
			} else if update.Message.Command() == "list" {
				cnf.Text = "subscriptions:"
				// @todo: why I used pointer?
				selectMenu = feedSelectMenu
				cnf.ReplyMarkup = PageMarkup(selectMenu)
			}

			if err != nil {
				panic(err)
			}

			c = cnf
		case update.CallbackQuery != nil && update.CallbackData() != "":
			// @todo: selectMenu initianalized to feedSelectMenu
			// selectMenu = feedSelectMenu

			if update.CallbackQuery.Data == "back" {
				// @todo: assign as pointer?
				selectMenu = feedSelectMenu
			} else if strings.HasPrefix(update.CallbackQuery.Data, "feed:") {
				index, err := strconv.Atoi(
					update.CallbackQuery.Data[len("feed:"):],
				)
				if err != nil {
					log.Panic(err)
				}
				itemSelectMenu.items = feedSelectMenu.feeds[index].Channel.Items
				itemSelectMenu.page = InitialPage
				selectMenu = itemSelectMenu
			} else if strings.HasPrefix(update.CallbackQuery.Data, "item:") {
				index, err := strconv.Atoi(
					update.CallbackQuery.Data[len("feed:"):],
				)
				if err != nil {
					log.Panic(err)
				}
				article := itemSelectMenu.items[index]
				toTelegraphNodeElement(article.Description.Value)
			}

			if update.CallbackQuery.Data == "next" {
				NextPage(selectMenu)
			} else if update.CallbackQuery.Data == "previous" {
				PrevPage(selectMenu)
			}

			c = tg.NewEditMessageReplyMarkup(
				update.CallbackQuery.Message.Chat.ID,
				update.CallbackQuery.Message.MessageID,
				PageMarkup(selectMenu),
			)
		default:
			c = tg.NewMessage(update.Message.Chat.ID, "Unknown command")
		}

		if _, err := bot.Send(c); err != nil {
			log.Panic(err)
		}
	}
}