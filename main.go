package main

import (
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"gopkg.in/ffmt.v1"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

//TODO bot de compra e vendas pro telegram
//start com botões bonitinhos pras ações do bot, o que fica na tela
//crud de produtos pra quem puder vender
//tem dono do bot
// cada produto tem:
//  -descrição
//  -preço
//aceitar crypto
//aceitar mandar contato do vendedor e uma mensagem pro vendedor
//aceitar apikey e owner na inicialização do bot
//guardar dados um uma db sqlite

var (
	owner  = os.Getenv("TELEGRAM_OWNER") // you should set this to your username
	ApiKey = os.Getenv("TELEGRAM_KEY")   // you should set this to your api key
	logger = log.Logger{}
	//todo add timeout on states
	//id -> state
	state   = map[string]string{}
	chans   = map[string]chan tgbotapi.Update{}
	catalog = map[string][]product{
		owner: []product{
			product{
				Description: "durr",
				Price:       666,
				Creator: tgbotapi.User{
					UserName: owner,
				},
			},
		},
	}
)

type product struct {
	Description string
	Price       float64
	Extra       []tgbotapi.Message //extra messages for selling the product
	Creator     tgbotapi.User
	OnBought    string
	CryptoKey   string
	Currency    string
	Done        bool
}

var mainMenuKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	//tgbotapi.NewInlineKeyboardRow(
	//	tgbotapi.NewInlineKeyboardButtonURL("1.com","http://1.com"),
	//	tgbotapi.NewInlineKeyboardButtonSwitch("2sw","open 2"),
	//	tgbotapi.NewInlineKeyboardButtonData("3","3"),
	//),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("List all sales", "/list"),
		tgbotapi.NewInlineKeyboardButtonData("Add sale", "/add"),
		tgbotapi.NewInlineKeyboardButtonData("Update sale", "/update"),
		tgbotapi.NewInlineKeyboardButtonData("Remove Sale", "/remove"),
	),
)

var exitKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Exit", "/exit"),
	),
)

var extnextKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Exit", "/exit"),
		tgbotapi.NewInlineKeyboardButtonData("Next", "/next"),
	),
)

//todo any other option?
var sendchargeKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Send contact to buyer", "/add/send_ctct"),
	),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Charge him crypto", "/add/charge_crypto"),
	),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Exit", "/exit"),
	),
)

type handler struct {
	Path    string
	Handler func(update tgbotapi.Update)
}

var handlers = []handler{}

func addHandler(path string, onmatch func(update tgbotapi.Update)) {
	handlers = append(handlers, handler{Path: path, Handler: onmatch})
}

//todo remove str != nil checks
func execHandlers(u tgbotapi.Update) {
	str := ""
	if u.Message != nil { //pvt message
		if str != "" {
			logger.Println("str was already used!")
		}
		str = u.Message.Text
	} else if u.ChannelPost != nil { //
		if str != "" {
			logger.Println("str was already used!")
		}
		str = u.ChannelPost.Text
	} else if u.ChosenInlineResult != nil { //
		if str != "" {
			logger.Println("str was already used!")
		}
		str = u.ChannelPost.Text
	} else if u.EditedChannelPost != nil {
		if str != "" {
			logger.Println("str was already used!")
		}
		str = u.EditedChannelPost.Text
	} else if u.EditedMessage != nil {
		if str != "" {
			logger.Println("str was already used!")
		}
		str = u.EditedMessage.Text
	} else if u.CallbackQuery != nil {
		if str != "" {
			logger.Println("str was already used!")
		}
		str = u.CallbackQuery.Data
	} else if u.InlineQuery != nil {
		if str != "" {
			logger.Println("str was already used!")
		}
		str = u.InlineQuery.Query
	} else if u.PreCheckoutQuery != nil {
		if str != "" {
			logger.Println("str was already used!")
		}
		str = u.PreCheckoutQuery.InvoicePayload
	} else if u.ShippingQuery != nil {
		if str != "" {
			logger.Println("str was already used!")
		}
		str = u.ShippingQuery.InvoicePayload
	}
	fmt.Print("str: ", str)
	for _, h := range handlers {
		match, err := regexp.MatchString(h.Path, str)
		if err != nil {
			logger.Print("error matching path to input string")
		}
		if match {
			h.Handler(u)
		}
	}

}

// Retrieves the user field "From" from every recieved update
func uFrom(u tgbotapi.Update) *tgbotapi.User {
	if u.CallbackQuery != nil {
		return u.CallbackQuery.From
	}
	if u.Message != nil {
		return u.Message.From
	}
	if u.ChannelPost != nil {
		return u.ChannelPost.From
	}
	if u.ShippingQuery != nil {
		return u.ShippingQuery.From
	}
	if u.PreCheckoutQuery != nil {
		return u.PreCheckoutQuery.From
	}
	if u.InlineQuery != nil {
		return u.InlineQuery.From
	}
	if u.EditedMessage != nil {
		return u.EditedMessage.From
	}
	if u.EditedChannelPost != nil {
		return u.EditedChannelPost.From
	}
	if u.ChosenInlineResult != nil {
		return u.ChosenInlineResult.From
	}
	return &tgbotapi.User{}
}

func uID(u tgbotapi.Update) int64 {
	if u.CallbackQuery != nil {
		return u.CallbackQuery.Message.Chat.ID
	}
	if u.Message != nil {
		return u.Message.Chat.ID
	}
	if u.ChannelPost != nil {
		return u.ChannelPost.Chat.ID
	}
	if u.ShippingQuery != nil {
		out, err := strconv.ParseInt(u.ShippingQuery.ID, 10, 64)
		if err != nil {
			fmt.Println("error converting string to int", u.ShippingQuery.ID, out, err)
		}
		return out
	}
	if u.PreCheckoutQuery != nil {
		out, err := strconv.ParseInt(u.PreCheckoutQuery.ID, 10, 64)
		if err != nil {
			fmt.Println("error converting string to int", u.PreCheckoutQuery.ID, out, err)
		}
		return out
	}
	if u.InlineQuery != nil {
		out, err := strconv.ParseInt(u.InlineQuery.ID, 10, 64)
		if err != nil {
			fmt.Println("error converting string to int", u.InlineQuery.ID, out, err)
		}
		return out
	}
	if u.EditedMessage != nil {
		return u.EditedMessage.Chat.ID
	}
	if u.EditedChannelPost != nil {
		return u.EditedChannelPost.Chat.ID
	}
	if u.ChosenInlineResult != nil {
		out, err := strconv.ParseInt(u.ChosenInlineResult.InlineMessageID, 10, 64)
		if err != nil {
			fmt.Println("error converting string to int", u.ChosenInlineResult.InlineMessageID, out, err)
		}
		return out
	}
	return 0
}

func main() {

	owner = os.Getenv("TELEGRAM_OWNER")
	ApiKey = os.Getenv("TELEGRAM_KEY")

	bot, err := tgbotapi.NewBotAPI(ApiKey)
	if err != nil {
		fmt.Print(ApiKey, owner)
		//log.Panic(err)
		return
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	addHandler("", func(update tgbotapi.Update) {
		fmt.Print("woo!\n")
	})

	addHandler("/list", func(update tgbotapi.Update) {
		if update.CallbackQuery != nil { //inline button press
			//bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data))
			for _, prod := range catalog {
				//bprod,err := json.Marshal(prod)
				//ffmt.Print(prod)
				//_ = bprod
				if err != nil {
					logger.Println("failure to marshall product into json")
				}
				msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, ffmt.Sprint(prod))
				msg.ReplyMarkup = exitKeyboard
				bot.Send(msg)
			}
		} else if update.Message != nil {
			for _, prod := range catalog {
				//bprod,err := json.Marshal(prod)
				//ffmt.Print(prod)
				//_ = bprod
				if err != nil {
					logger.Println("failure to marshall product into json")
				}
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, ffmt.Sprint(prod))
				msg.ReplyMarkup = exitKeyboard
				bot.Send(msg)
			}
		}
	})

	addHandler("/add", func(update tgbotapi.Update) {
		//if update.CallbackQuery != nil { //inline button press
		//bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data))
		msg := tgbotapi.NewMessage(uID(update), "Send the product description")
		msg.ReplyMarkup = exitKeyboard
		bot.Send(msg)
		//wait for answer

		state[uFrom(update).UserName] = "add_desc"
		update = <-chans[uFrom(update).UserName]

		catalog[update.Message.From.UserName] = append(catalog[update.Message.From.UserName], product{Description: update.Message.Text})
		msg = tgbotapi.NewMessage(uID(update), "Send the product price")
		msg.ReplyMarkup = exitKeyboard
		bot.Send(msg)

		state[uFrom(update).UserName] = "add_price"
		update = <-chans[uFrom(update).UserName]

		f, err := strconv.ParseFloat(update.Message.Text, 64)
		if err != nil {
			msg := tgbotapi.NewMessage(uID(update), "Only send a number, please")
			msg.ReplyMarkup = exitKeyboard
			bot.Send(msg)
			//logger.Print("unable to convert message to float")
		} else {
			catalog[update.Message.From.UserName][len(catalog[update.Message.From.UserName])-1].Price = f
			msg := tgbotapi.NewMessage(uID(update), "Send any other message you'd like the buyer to receive, then press the next button")
			msg.ReplyMarkup = extnextKeyboard
			bot.Send(msg)
		}

		state[uFrom(update).UserName] = "add_other"

		for { //recieve until next is pressed
			update = <-chans[uFrom(update).UserName]
			if update.CallbackQuery != nil && update.CallbackQuery.Data == "/next" { //inline button press
				bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data))
				msg := tgbotapi.NewMessage(uID(update), "Chose what to do when an item is bought")
				msg.ReplyMarkup = sendchargeKeyboard
				bot.Send(msg)
				break
			} else if update.CallbackQuery != nil && update.CallbackQuery.Data == "/exit" {
				bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data))
				if len(catalog[update.CallbackQuery.From.UserName]) > 0 && !catalog[update.CallbackQuery.From.UserName][len(catalog[update.CallbackQuery.From.UserName])-1].Done {
					catalog[update.CallbackQuery.From.UserName] = catalog[update.CallbackQuery.From.UserName][:len(catalog[update.CallbackQuery.From.UserName])-1]
				}
				state[uFrom(update).UserName] = ""
				return
			} else {
				catalog[uFrom(update).UserName][len(catalog[uFrom(update).UserName])-1].Extra = append(catalog[uFrom(update).UserName][len(catalog[uFrom(update).UserName])-1].Extra, *update.Message)
			}
		}

		state[uFrom(update).UserName] = "add_sendorcharge"
		msg = tgbotapi.NewMessage(uID(update), "Please answer what to do when the user buys the product")
		msg.ReplyMarkup = sendchargeKeyboard

		for {
			update = <-chans[uFrom(update).UserName]
			if update.CallbackQuery != nil && update.CallbackQuery.Data == "/add/charge_crypto" {
				bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data))
				catalog[uFrom(update).UserName][len(catalog[uFrom(update).UserName])-1].OnBought = "crypto"
				bot.Send(tgbotapi.NewMessage(uID(update), "Please state the favored cryptocurrency short name, e.g. BTC ETH XMR"))
				break
			} else if update.CallbackQuery != nil && update.CallbackQuery.Data == "/add/send_ctct" {
				bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data))
				catalog[uFrom(update).UserName][len(catalog[uFrom(update).UserName])-1].OnBought = "contact"
				msg := tgbotapi.NewMessage(uID(update), "Thank you for adding a product for sale!")
				msg.ReplyMarkup = mainMenuKeyboard
				bot.Send(msg)
				state[uFrom(update).UserName] = ""
				return
			} else {
				msg := tgbotapi.NewMessage(uID(update), "Please choose an item")
				msg.ReplyMarkup = sendchargeKeyboard
				bot.Send(msg)
			}
		}

		state[uFrom(update).UserName] = "add_get_crypto"
		update = <-chans[uFrom(update).UserName]
		catalog[update.Message.From.UserName][len(catalog[update.Message.From.UserName])-1].Currency = strings.Trim(update.Message.Text, " .,\n")
		msg = tgbotapi.NewMessage(uID(update), "Please send your public key")
		bot.Send(msg)

		state[uFrom(update).UserName] = "add_get_crypto_key"
		update = <-chans[uFrom(update).UserName]
		catalog[update.Message.From.UserName][len(catalog[update.Message.From.UserName])-1].CryptoKey = strings.Trim(update.Message.Text, " .,\n")
		catalog[update.Message.From.UserName][len(catalog[update.Message.From.UserName])-1].Done = true
		msg = tgbotapi.NewMessage(uID(update), "Thank you for adding a product for sale!")
		msg.ReplyMarkup = mainMenuKeyboard
		bot.Send(msg)
		state[uFrom(update).UserName] = ""
		return
	})

	addHandler("", func(update tgbotapi.Update) {
		if update.CallbackQuery != nil { //inline button press
			bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data))
			//fmt.Print("callback data: ", update.CallbackQuery.Data, "\n")
		}
		if update.Message != nil {
			switch state[uFrom(update).UserName] {
			// TODO: add exit inline keyboard on the add states
			default:
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "fdsfgs")
				switch update.Message.Text {
				case "/start":
					msg.ReplyMarkup = mainMenuKeyboard
				}
				if update.Message.Text == "/start@"+bot.Self.UserName {
					msg.ReplyMarkup = mainMenuKeyboard
				}
				bot.Send(msg)
			}
		}
	})

	time.Sleep(time.Millisecond * 500)
	updates.Clear()

	for update := range updates {
		if chans[uFrom(update).UserName] == nil {
			chans[uFrom(update).UserName] = make(chan tgbotapi.Update, 16)

			go func() {
				from := uFrom(update).UserName
				tout := time.After(30 * time.Minute)
				last := time.Now()
				for {
					select {
					case u := <-chans[from]:
						last = time.Now()
						execHandlers(u)
					case <-tout:
						//no longer active
						if len(chans[from]) == 0 && time.Now().Sub(last) > 17*time.Minute {
							close(chans[from])
							return
						} else {
							tout = time.After(time.Hour)
						}
					}
				}
			}()
		}
		chans[uFrom(update).UserName] <- update
		// if nil make chan and goroutine
		// send to chan
		//
		//execHandlers(&update)
		//ffmt.Print(update)
	}
}
