package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/growthbook/growthbook-golang"
)

type GrowthBookResp struct {
	Features json.RawMessage
	Status   int
}

func NewGrowthBook(features growthbook.FeatureMap, attributeID string) *growthbook.GrowthBook {
	context := growthbook.NewContext().
		WithFeatures(features).
		WithAttributes(growthbook.Attributes{"id": attributeID})
	return growthbook.New(context)
}

func RunExperiment(features growthbook.FeatureMap, experimentName string, attributeID string) (*growthbook.ExperimentResult, string, error) {
	if features == nil {
		err := fmt.Errorf("features is nil")
		return nil, "", err
	}
	if features[experimentName] == nil {
		err := fmt.Errorf("experiment %s not found", experimentName)
		return nil, "", err
	}
	if len(features[experimentName].Rules) < 1 {
		err := fmt.Errorf("experiment %s has no rules", experimentName)
		return nil, "", err
	}
	var rule *growthbook.FeatureRule
	for _, item := range features[experimentName].Rules {
		if *item.TrackingKey == experimentName {
			rule = item
			break
		}
	}
	if rule == nil {
		return nil, "", fmt.Errorf("experiment rule not found")
	}
	gbInstance := NewGrowthBook(features, attributeID)
	variations := rule.Variations
	weights := rule.Weights
	coverage := float64(*rule.Coverage)
	namespace := rule.Namespace
	exp := growthbook.NewExperiment(experimentName).
		WithVariations(variations...).
		WithWeights(weights...).
		WithCoverage(coverage).
		WithNamespace(namespace)
	var namespaceStr string
	if namespace != nil {
		namespaceStr = namespace.ID
	}
	return gbInstance.Run(exp), namespaceStr, nil
}

func GetGrowthbookFeatures() (growthbook.FeatureMap, error) {

	resp, err := http.Get(os.Getenv("GROWTHBOOK_API_URL"))
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var gbResp GrowthBookResp
	err = json.Unmarshal(body, &gbResp)
	if err != nil {
		return nil, err
	}
	isDebug, _ := strconv.ParseBool(os.Getenv("DEBUG"))
	if isDebug {
		featuresJson, _ := json.Marshal(gbResp)
		log.Println(string(featuresJson))
	}

	return growthbook.ParseFeatureMap(gbResp.Features), nil
}
