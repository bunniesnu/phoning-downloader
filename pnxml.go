package main

import (
	"encoding/json"
	"fmt"
	"strconv"
)

func getPNXML(apiKey, accessToken string, id int) (map[string]any, error) {
	endpoint := "/fan/v1.0/lives/" + strconv.Itoa(id) + "/play-info-v3"
	params := map[string]string{
		"countryCode": "KR",
	}
	res, err := phoning(apiKey, accessToken, endpoint, params)
	if err != nil {
		return res, err
	}

    raw, ok := res["data"].(map[string]any)["lipPlayback"]
    if !ok {
        return nil, fmt.Errorf("missing lipPlayback field")
    }
    lipJSON, ok := raw.(string)
    if !ok {
        return nil, fmt.Errorf("lipPlayback is not a JSON string (got %T)", raw)
    }

    var lipMap map[string]any
    if err := json.Unmarshal([]byte(lipJSON), &lipMap); err != nil {
        return nil, fmt.Errorf("failed to parse lipPlayback JSON: %w", err)
    }

    period, ok := lipMap["period"].([]any)
    if !ok {
        return nil, fmt.Errorf("lipPlayback.period not present")
    }
	if len(period) != 1 {
		return nil, fmt.Errorf("lipPlayback.period should have exactly one element, got %d", len(period))
	}
	periodMap, ok := period[0].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("lipPlayback.period is not a map")
	}
	adaptationSet, ok := periodMap["adaptationSet"].([]any)
	if !ok {
		return nil, fmt.Errorf("lipPlayback.period.adaptationSet not present")
	}
	if len(adaptationSet) == 0 {
		return nil, fmt.Errorf("lipPlayback.period.adaptationSet is empty")
	}
	adaptationSetMap, ok := adaptationSet[0].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("lipPlayback.period.adaptationSet is not a map")
	}
	maxWidth, ok := adaptationSetMap["maxWidth"].(float64)
	if !ok {
		return nil, fmt.Errorf("lipPlayback.period.adaptationSet.maxWidth not present")
	}
	representation, ok := adaptationSetMap["representation"].([]any)
	if !ok {
		return nil, fmt.Errorf("lipPlayback.period.adaptationSet.representation not present")
	}
	if len(representation) == 0 {
		return nil, fmt.Errorf("lipPlayback.period.adaptationSet.representation is empty")
	}
	for _, rep := range representation {
		repMap, ok := rep.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("lipPlayback.period.adaptationSet.representation is not a map")
		}
		if repMap["width"] == maxWidth {
			baseURL, ok := repMap["baseURL"].([]any)
			if !ok {
				return nil, fmt.Errorf("lipPlayback.period.adaptationSet.representation.baseURL not present")
			}
			if len(baseURL) == 0 {
				return nil, fmt.Errorf("lipPlayback.period.adaptationSet.representation.baseURL is empty")
			}
			baseURLMap, ok := baseURL[0].(map[string]any)
			if !ok {
				return nil, fmt.Errorf("lipPlayback.period.adaptationSet.representation.baseURL is not a map")
			}
			baseURLvalue, ok := baseURLMap["value"].(string)
			if !ok {
				return nil, fmt.Errorf("lipPlayback.period.adaptationSet.representation.baseURL.value not present")
			}
			returnValue := make(map[string]any)
			returnValue["url"] = baseURLvalue
			return returnValue, nil
		}
	}
	return nil, fmt.Errorf("no suitable representation found in lipPlayback")
}