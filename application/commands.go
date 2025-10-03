package application

import (
	"fmt"

	"github.com/dmpettyp/artwork/domain/imagegraph"
	"github.com/dmpettyp/dorky"
)

type CreateImageGraphCommand struct {
	dorky.BaseCommand
	ImageGraphID imagegraph.ImageGraphID
	Name         string
}

func NewCreateImageGraphCommand(
	name string,
) (*CreateImageGraphCommand, error) {
	imageGraphID, err := imagegraph.NewImageGraphID()

	if err != nil {
		return nil, fmt.Errorf(
			"could not create new CreateImageGraphCommand: %w",
			err,
		)
	}
	command := &CreateImageGraphCommand{
		ImageGraphID: imageGraphID,
		Name:         name,
	}

	err = command.Init("CreateImageGraphCommand")

	if err != nil {
		return nil, fmt.Errorf(
			"could not create new CreateImageGraphCommand: %w",
			err,
		)
	}

	return command, nil
}
