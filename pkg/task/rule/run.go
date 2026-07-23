package rule

import (
	"github.com/sirupsen/logrus"

)

func (r *RuleTask) Run() ([]any, string, error) {
	r.logger.WithFields(logrus.Fields{
		"source":      r.source,
		"destination": r.destination,
	}).Debug("generate sync plan")

	sourceURLs, destinationURLs, err := r.generateRepoURLs()
	if err != nil {
		return nil, "", err
	}

	results := make([]any, 0, len(sourceURLs))

	for index, sourceURL := range sourceURLs {
		destinationURL := destinationURLs[index]

		if r.plan != nil {
			r.plan.AddImage(
				sourceURL.String(),
				destinationURL.String(),
				sourceURL.GetURLWithoutTagOrDigest(),
				destinationURL.GetURLWithoutTagOrDigest(),
			)
		}

		r.logger.WithFields(logrus.Fields{
			"source":      sourceURL.String(),
			"destination": destinationURL.String(),
		}).Info("plan image sync")

		t, err := r.BuildURLTask(sourceURL, destinationURL, r.urlTaskFactory)
		if err != nil {
			return nil, "", err
		}

		results = append(results, t)
	}

	return results, "sync plan generated", nil
}

