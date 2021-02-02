var graph_module = require("@arangodb/general-graph");

var edgeDefinitions = [ { collection: "edge_verif_scheme", "from": [ "hwid_collection" ], "to" : [ "swid_collection" ] } ];
var edgePatchDefinitions = [ { collection: "edge_rel_scheme", "from": [ "swid_collection" ], "to" : [ "swid_collection" ] } ];
var graph = graph_module._create("psa-endorsements", edgeDefinitions);
var graphPatch = graph_module._create("psa-patch-endorsements", edgePatchDefinitions);
