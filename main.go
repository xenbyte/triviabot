package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	EmbededMessages "quizbot/embededMessages"
	widgets "quizbot/reactions"
	"strconv"
	"syscall"
	"time"
)

var Token string

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()
}

type Question struct {
	ResponseCode int `json:"response_code"`
	Results      []struct {
		Category         string   `json:"category"`
		Type             string   `json:"type"`
		Difficulty       string   `json:"difficulty"`
		Question         string   `json:"question"`
		CorrectAnswer    string   `json:"correct_answer"`
		IncorrectAnswers []string `json:"incorrect_answers"`
	} `json:"results"`
}

func main() {

	if Token == "" {
		fmt.Println("No Token Provided. Please Run quizbot -t <bot token>")
	}

	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("Error Creating a discord session, ", err)
	}
	dg.AddHandler(ready)
	dg.AddHandler(messageCreate)

	err = dg.Open()
	if err != nil {
		fmt.Println("Error opening Discord session: ", err)
	}
	fmt.Println("The bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Kill, os.Interrupt)
	<-sc
}

func ready(s *discordgo.Session, event *discordgo.Event) {
	s.UpdateGameStatus(0, "!help for info")
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	if m.Content == "!help" {
		_, err := s.ChannelMessageSendEmbed(m.ChannelID, EmbededMessages.NewGenericsEmbed(
			"Help menu for QuizBot",
			fmt.Sprintf("Welcome to the QuizBot.\nBelow are the commands to get started.\n\n"+
				"**!ask:\t\t **asks a question\n\n"+
				"**!score:\t\t **watch current leaderboard.\n\n\n"),
			"a bot by hsnb3h.",
		))
		if err != nil {
			fmt.Println("Error occurred in sending message: ", err)
		}

	} else if m.Content == "!ask" {
		questionUrl := "https://opentdb.com/api.php?amount=1&category=18&type=multiple"
		var question Question
		client := http.Client{}
		req, err := http.NewRequest("GET", questionUrl, nil)
		if err != nil {
			fmt.Println("Error Sending GET request to opentdb.")
		}
		res, err := client.Do(req)
		if err != nil {
			return
		}
		body, err := ioutil.ReadAll(res.Body)
		if err := json.Unmarshal(body, &question); err != nil {
			fmt.Println("Error happened in Unmarshall JSON: ", err)
		}
		questionTBA := question.Results[0].Question
		correctAnswer := question.Results[0].CorrectAnswer
		allAnswers := make([]string, 4)
		optionWidgets := make([]string, 4)
		allAnswers[0] = question.Results[0].IncorrectAnswers[0]
		allAnswers[1] = question.Results[0].IncorrectAnswers[1]
		allAnswers[2] = question.Results[0].IncorrectAnswers[2]
		allAnswers[3] = correctAnswer
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(allAnswers), func(i, j int) { allAnswers[i], allAnswers[j] = allAnswers[j], allAnswers[i] })
		optionWidgets[0] = widgets.OptionOne
		optionWidgets[1] = widgets.OptionTwo
		optionWidgets[2] = widgets.OptionThree
		optionWidgets[3] = widgets.OptionFour
		message, err := s.ChannelMessageSendEmbed(m.ChannelID, EmbededMessages.NewGenericsEmbed(
			"Question #1",
			fmt.Sprintf("%s\n\n"+
				"1️⃣  %s\n"+
				"2️⃣  %s\n"+
				"3️⃣  %s\n"+
				"4️⃣  %s\n", questionTBA, allAnswers[0], allAnswers[1], allAnswers[2], allAnswers[3]),
			"",
		))
		if err != nil {
			return
		}
		for _, opt := range optionWidgets {
			err := s.MessageReactionAdd(m.ChannelID, message.ID, opt)
			if err != nil {
				return
			}
		}
		s.MessageReactionAdd(m.ChannelID, message.ID, ":ballot_box_with_check:")
		s.AddHandler(func(s *discordgo.Session, m *discordgo.MessageReactionAdd) {
			func() {
				qIndexStr := m.Emoji.Name[0:1]
				qIndex, _ := strconv.Atoi(qIndexStr)
				if allAnswers[qIndex-1] == correctAnswer && correctAnswer != "" && len(allAnswers) != 0 {
					fmt.Println("Question: ", questionTBA)
					fmt.Println("Correct Answer!")
					allAnswers = make([]string, 4)
					correctAnswer = ""
					s.ChannelMessageDelete(m.ChannelID, m.MessageID)
					message, err := s.ChannelMessageSendEmbed(m.ChannelID, EmbededMessages.NewGenericsEmbed("Correct!", "You've Correctly Answered The Question!", ""))
					if err != nil {
						fmt.Println("Error")
					}
					fmt.Println(message)
				} else if allAnswers[qIndex-1] != correctAnswer && correctAnswer != "" && len(allAnswers) != 0 {
					fmt.Println("Question: ", questionTBA)
					fmt.Println("All Answers: ", allAnswers)
					fmt.Println("Incorrect Answer!")
					allAnswers = make([]string, 4)
					correctAnswer = ""
					s.ChannelMessageDelete(m.ChannelID, m.MessageID)
					message, err := s.ChannelMessageSendEmbed(m.ChannelID, EmbededMessages.NewGenericsEmbed("Incorrect!", "You've Incorrectly Answered The Question!", ""))
					if err != nil {
						fmt.Println("Error")
					}
					fmt.Println(message)
				}
			}()
		})
	}
}

func messageReactions(s *discordgo.Session, reactions *discordgo.MessageReactions) {
	fmt.Println(reactions.Emoji.User.ID)
}
