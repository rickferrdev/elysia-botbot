package interactionflow

import (
	"context"
	"time"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/components"
	"github.com/discord-go/discord.go/rest"
	"go.uber.org/fx"
)

var Provide = fx.Provide(fx.Annotate(New, fx.As(new(Interface))))

type (
	Service struct{}
)

type Interface interface {
	AwaitStringSelect(interaction *bot.InteractionContext, options []components.SelectOption, selectMenu components.StringSelect) (*bot.InteractionContext, error)
	AwaitSimpleTextModal(interaction *bot.InteractionContext, textInput components.TextInput, modalData components.ModalData) (*bot.InteractionContext, error)
	AwaitChannelSelect(interaction *bot.InteractionContext, selectMenu components.ChannelSelect) (*bot.InteractionContext, error)
}

func New() *Service {
	return &Service{}
}

func (service *Service) AwaitStringSelect(interaction *bot.InteractionContext, options []components.SelectOption, selectMenu components.StringSelect) (*bot.InteractionContext, error) {
	selectMenu.Options = options
	row := components.NewActionRowBuilder().AddComponents(selectMenu).Build()
	if _, err := interaction.EditReplyComplex(rest.EditMessageParams{
		Components: &[]components.Component{row},
	}); err != nil {
		return nil, err
	}

	timeout, cancel := context.WithTimeout(interaction.Context(), 2*time.Minute)
	defer cancel()
	selected, err := interaction.Bot.AwaitInteraction(timeout, func(ic *bot.InteractionContext) bool {
		return ic.CustomID() == selectMenu.CustomID &&
			ic.User().ID == interaction.User().ID
	})
	if err != nil {
		return nil, err
	}

	return selected, nil
}

func (service *Service) AwaitSimpleTextModal(interaction *bot.InteractionContext, textInput components.TextInput, modalData components.ModalData) (*bot.InteractionContext, error) {
	row := components.NewActionRowBuilder().AddComponents(textInput).Build()
	modal := components.NewModalBuilder().
		SetTitle(modalData.Title).
		SetCustomID(modalData.CustomID).
		AddComponents(row).
		Build()

	if err := interaction.ShowModalBuilder(modal); err != nil {
		return nil, err
	}

	timeout, cancel := context.WithTimeout(interaction.Context(), 2*time.Minute)
	defer cancel()
	submission, err := interaction.Bot.AwaitInteraction(timeout, func(ic *bot.InteractionContext) bool {
		return ic.IsModalSubmit() &&
			ic.CustomID() == modalData.CustomID &&
			ic.User().ID == interaction.User().ID
	})
	if err != nil {
		return nil, err
	}

	return submission, nil
}

func (service *Service) AwaitChannelSelect(interaction *bot.InteractionContext, selectMenu components.ChannelSelect) (*bot.InteractionContext, error) {
	row := components.NewActionRowBuilder().AddComponents(selectMenu).Build()
	if _, err := interaction.EditReplyComplex(rest.EditMessageParams{
		Components: &[]components.Component{row},
	}); err != nil {
		return nil, err
	}

	timeout, cancel := context.WithTimeout(interaction.Context(), 2*time.Minute)
	defer cancel()
	selected, err := interaction.Bot.AwaitInteraction(timeout, func(ic *bot.InteractionContext) bool {
		return ic.CustomID() == selectMenu.CustomID &&
			ic.User().ID == interaction.User().ID
	})
	if err != nil {
		return nil, err
	}

	return selected, nil
}
