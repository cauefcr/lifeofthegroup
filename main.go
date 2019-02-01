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
	state   = map[int64]string{}
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
		tgbotapi.NewInlineKeyboardButtonData("List all sales", "lst"),
		tgbotapi.NewInlineKeyboardButtonData("Add sale", "add"),
		tgbotapi.NewInlineKeyboardButtonData("Update sale", "upd"),
		tgbotapi.NewInlineKeyboardButtonData("Remove Sale", "rem"),
	),
)

var exitKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Exit", "ext"),
	),
)

var extnextKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Exit", "ext"),
		tgbotapi.NewInlineKeyboardButtonData("Next", "nxt"),
	),
)

//todo any other option?
var sendchargeKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Send contact to buyer", "send_ctct"),
	),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Charge him crypto", "charge_crypto"),
	),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Exit", "ext"),
	),
)

type handler struct {
	Path    string
	Handler func(update *tgbotapi.Update)
}

var handlers = []handler{}

func addHandler(path string, onmatch func(update *tgbotapi.Update)) {
	handlers = append(handlers, handler{Path: path, Handler: onmatch})
}

//todo remove str != nil checks
func execHandlers(u *tgbotapi.Update) {
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
	addHandler("", func(update *tgbotapi.Update) {
		fmt.Print("woo!\n")
	})
	time.Sleep(time.Millisecond * 500)
	updates.Clear()

	for update := range updates {
		execHandlers(&update)
		//if button press
		//ffmt.Print(update)
		if update.CallbackQuery != nil {
			bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data))

			switch update.CallbackQuery.Data {
			case "lst":
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
				break
			case "add":
				state[update.CallbackQuery.Message.Chat.ID] = "add_desc"
				msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Send the product description")
				msg.ReplyMarkup = exitKeyboard
				bot.Send(msg)
				break
			case "upd":

				break
			case "rem":

				break
			case "ext":
				//actually delete the half-made product
				if len(catalog[update.CallbackQuery.From.UserName]) > 0 && !catalog[update.CallbackQuery.From.UserName][len(catalog[update.CallbackQuery.From.UserName])-1].Done {
					catalog[update.CallbackQuery.From.UserName] = catalog[update.CallbackQuery.From.UserName][:len(catalog[update.CallbackQuery.From.UserName])-1]
				}
				state[update.CallbackQuery.Message.Chat.ID] = ""
			case "nxt":
				if state[update.CallbackQuery.Message.Chat.ID] == "add_other" {
					state[update.CallbackQuery.Message.Chat.ID] = "add_sendorcharge"
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Chose what to do when an item is bought")
					msg.ReplyMarkup = sendchargeKeyboard
					bot.Send(msg)
				}
			case "charge_crypto":
				catalog[update.CallbackQuery.Message.From.UserName][len(catalog[update.CallbackQuery.Message.From.UserName])-1].OnBought = "crypto"
				bot.Send(tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Please state the favored cryptocurrency short name, e.g. BTC ETH XMR"))
				state[update.CallbackQuery.Message.Chat.ID] = "add_get_crypto"
				break
			case "send_ctct":
				catalog[update.CallbackQuery.Message.From.UserName][len(catalog[update.CallbackQuery.Message.From.UserName])-1].OnBought = "contact"
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Thank you for adding a product for sale!")
				msg.ReplyMarkup = mainMenuKeyboard
				bot.Send(msg)
				state[update.CallbackQuery.Message.Chat.ID] = ""
				break
			default:
				logger.Print("callback data: ", update.CallbackQuery.Data, "\n")
			}

			bot.Send(tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Data))
		}
		//if regular message
		if update.Message != nil {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "fdsfgs")

			switch state[update.Message.Chat.ID] {

			// TODO: add exit inline keyboard on the add states
			case "add_desc":
				catalog[update.Message.From.UserName] = append(catalog[update.Message.From.UserName], product{Description: update.Message.Text})
				state[update.Message.Chat.ID] = "add_price"
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Send the product price")
				msg.ReplyMarkup = exitKeyboard
				bot.Send(msg)
			case "add_price":
				f, err := strconv.ParseFloat(update.Message.Text, 64)
				if err != nil {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Only send a number, please")
					msg.ReplyMarkup = exitKeyboard
					bot.Send(msg)
					//logger.Print("unable to convert message to float")
				} else {
					catalog[update.Message.From.UserName][len(catalog[update.Message.From.UserName])-1].Price = f
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Send any other message you'd like the buyer to receive, then press the next button")
					msg.ReplyMarkup = extnextKeyboard
					bot.Send(msg)
				}
				state[update.Message.Chat.ID] = "add_other"
			case "add_other":
				catalog[update.Message.From.UserName][len(catalog[update.Message.From.UserName])-1].Extra = append(catalog[update.Message.From.UserName][len(catalog[update.Message.From.UserName])-1].Extra, *update.Message)
			case "add_sendorcharge":
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Please answer what to do when the user buys the product")
				msg.ReplyMarkup = sendchargeKeyboard
			case "add_get_crypto":
				catalog[update.Message.From.UserName][len(catalog[update.Message.From.UserName])-1].Currency = strings.Trim(update.Message.Text, " .,\n")
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Please send your public key")
				bot.Send(msg)
				state[update.Message.Chat.ID] = "add_get_crypto_key"
			case "add_get_crypto_key":
				catalog[update.Message.From.UserName][len(catalog[update.Message.From.UserName])-1].CryptoKey = strings.Trim(update.Message.Text, " .,\n")
				catalog[update.Message.From.UserName][len(catalog[update.Message.From.UserName])-1].Done = true
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Thank you for adding a product for sale!")
				msg.ReplyMarkup = mainMenuKeyboard
				bot.Send(msg)
				state[update.Message.Chat.ID] = ""
			default:
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
	}
}
