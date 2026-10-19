package environment

import "strings"

// These names mirror the bundled m365 table catalog. The catalog test detects
// changes so compatibility aliases remain synchronized without coupling the
// language layer to the application schema assets.
var defenderTables = map[string]bool{
	"aadsignineventsbeta":                      true,
	"aadspnsignineventsbeta":                   true,
	"agentsinfo":                               true,
	"aiagentsinfo":                             true,
	"alertevidence":                            true,
	"alertinfo":                                true,
	"behaviorentities":                         true,
	"behaviorinfo":                             true,
	"campaigninfo":                             true,
	"cloudappevents":                           true,
	"cloudauditevents":                         true,
	"clouddnsevents":                           true,
	"cloudpolicyenforcementevents":             true,
	"cloudprocessevents":                       true,
	"cloudstorageaggregatedevents":             true,
	"datasecuritybehaviors":                    true,
	"datasecurityevents":                       true,
	"devicebaselinecomplianceassessment":       true,
	"devicebaselinecomplianceassessmentkb":     true,
	"devicebaselinecomplianceprofiles":         true,
	"deviceevents":                             true,
	"devicefilecertificateinfo":                true,
	"devicefileevents":                         true,
	"deviceimageloadevents":                    true,
	"deviceinfo":                               true,
	"devicelogonevents":                        true,
	"devicenetworkevents":                      true,
	"devicenetworkinfo":                        true,
	"deviceprocessevents":                      true,
	"deviceregistryevents":                     true,
	"devicetvmbrowserextensions":               true,
	"devicetvmbrowserextensionskb":             true,
	"devicetvmcertificateinfo":                 true,
	"devicetvmhardwarefirmware":                true,
	"devicetvminfogathering":                   true,
	"devicetvminfogatheringkb":                 true,
	"devicetvmsecureconfigurationassessment":   true,
	"devicetvmsecureconfigurationassessmentkb": true,
	"devicetvmsoftwareevidencebeta":            true,
	"devicetvmsoftwareinventory":               true,
	"devicetvmsoftwarevulnerabilities":         true,
	"devicetvmsoftwarevulnerabilitieskb":       true,
	"disruptionandresponseevents":              true,
	"emailattachmentinfo":                      true,
	"emailevents":                              true,
	"emailpostdeliveryevents":                  true,
	"emailurlinfo":                             true,
	"entraidsigninevents":                      true,
	"entraidspnsigninevents":                   true,
	"exposuregraphedges":                       true,
	"exposuregraphnodes":                       true,
	"filemaliciouscontentinfo":                 true,
	"graphapiauditevents":                      true,
	"identityaccountinfo":                      true,
	"identitydirectoryevents":                  true,
	"identityevents":                           true,
	"identityinfo":                             true,
	"identitylogonevents":                      true,
	"identityqueryevents":                      true,
	"messageevents":                            true,
	"messagepostdeliveryevents":                true,
	"messageurlinfo":                           true,
	"oauthappinfo":                             true,
	"urlclickevents":                           true,
}

// Missing aliases inherit the declared time column's type. If neither name is
// declared, expose both as datetime. Explicit types supplied by users are kept.
func allowDefenderTimeAliases(tables map[string]TableDetails) {
	for name, columns := range tables {
		if !defenderTables[strings.ToLower(name)] {
			continue
		}
		if columns == nil {
			columns = TableDetails{}
			tables[name] = columns
		}
		timestampType, hasTimestamp := timeColumnType(columns, "Timestamp")
		generatedType, hasGenerated := timeColumnType(columns, "TimeGenerated")
		switch {
		case !hasTimestamp && !hasGenerated:
			columns["Timestamp"], columns["TimeGenerated"] = "datetime", "datetime"
		case !hasTimestamp:
			columns["Timestamp"] = generatedType
		case !hasGenerated:
			columns["TimeGenerated"] = timestampType
		}
	}
}

func timeColumnType(columns TableDetails, name string) (string, bool) {
	if typ, exists := columns[name]; exists {
		return typ, true
	}
	for column, typ := range columns {
		if strings.EqualFold(column, name) {
			return typ, true
		}
	}
	return "", false
}
