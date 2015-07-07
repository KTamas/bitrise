package cli

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise-cli/bitrise"
	"github.com/bitrise-io/bitrise-cli/models"
	"github.com/bitrise-io/go-pathutil"
	"github.com/codegangsta/cli"
)

func activateAndRunSteps(workflow models.WorkflowModel) error {
	for _, step := range workflow.Steps {
		stepDir := "./steps/" + step.Id + "/" + step.VersionTag + "/"

		if err := bitrise.RunStepmanActivate(step.Id, step.VersionTag, stepDir); err != nil {
			log.Errorln("[BITRISE_CLI] - Failed to run stepman activate")
			return err
		}

		log.Infof("[BITRISE_CLI] - Step activated: %s (%s)", step.Id, step.VersionTag)

		if err := runStep(step); err != nil {
			log.Errorln("[BITRISE_CLI] - Failed to run step")
			return err
		}
	}
	return nil
}

func runStep(step models.StepModel) error {
	// Add step envs
	for _, input := range step.Inputs {
		if input.Value != nil {
			if err := bitrise.RunEnvmanAdd(*input.MappedTo, *input.Value); err != nil {
				log.Errorln("[BITRISE_CLI] - Failed to run envman add")
				return err
			}
		}
	}

	stepDir := "./steps/" + step.Id + "/" + step.VersionTag + "/"
	stepCmd := fmt.Sprintf("%sstep.sh", stepDir)
	cmd := []string{"bash", stepCmd}

	if err := bitrise.RunEnvmanRun(cmd); err != nil {
		log.Errorln("[BITRISE_CLI] - Failed to run envman run")
		return err
	}

	log.Infof("[BITRISE_CLI] - Step executed: %s (%s)", step.Id, step.VersionTag)
	return nil
}

func doRun(c *cli.Context) {
	log.Info("[BITRISE_CLI] - Run")

	// Input validation
	workflowJsonPath := c.String(PATH_KEY)
	if workflowJsonPath == "" {
		log.Infoln("[BITRISE_CLI] - Workflow json path not defined, try search in current folder")

		if exist, err := pathutil.IsPathExists("./bitrise.json"); err != nil {
			log.Fatalln("Failed to check path:", err)
		} else if !exist {
			log.Fatalln("No workflow json found")
		}
		workflowJsonPath = "./bitrise.json"
	}

	// Envman setup
	if err := os.Setenv(ENVSTORE_PATH_ENV_KEY, ENVSTORE_PATH); err != nil {
		log.Fatalln("[BITRISE_CLI] - Failed to add env:", err)
	}

	if err := os.Setenv(FORMATTED_OUTPUT_PATH_ENV_KEY, FORMATTED_OUTPUT_PATH); err != nil {
		log.Fatalln("[BITRISE_CLI] - Failed to add env:", err)
	}

	if err := bitrise.RunEnvmanInit(); err != nil {
		log.Fatalln("[BITRISE_CLI] - Failed to run envman init")
	}

	// Run work flow
	if workflow, err := bitrise.ReadWorkflowJson(workflowJsonPath); err != nil {
		log.Fatalln("[BITRISE_CLI] - Failed to read work flow:", err)
	} else {
		if err := activateAndRunSteps(workflow); err != nil {
			log.Fatalln("[BITRISE_CLI] - Failed to activate steps:", err)
		}
	}
}