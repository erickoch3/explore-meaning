package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/ekoch/explore_meaning/internal/graph"
	"github.com/ekoch/explore_meaning/internal/semantics"
)

var wordGraph *graph.WordGraph

// Define a helper function to increment index for 1-based numbering
var funcMap = template.FuncMap{
	"inc": func(i int) int {
		return i + 1
	},
}

// LemmaRatio represents a lemma and its ratio for analysis
type LemmaRatio struct {
	Lemma     semantics.Lemma
	DefinedBy int
	Defines   int
	Ratio     float64
	WordCount int
	Category  string // "producer", "consumer", or "balanced"
	IsExtreme bool   // Whether it's an extreme case (only defines/defined-by)
}

var templates = template.Must(template.New("").Funcs(funcMap).Parse(`
<!DOCTYPE html>
<html>
<head>
    <title>Word Graph Explorer</title>
    <script src="https://d3js.org/d3.v7.min.js"></script>
    <style>
        body {
            font-family: Arial, sans-serif;
            line-height: 1.6;
            margin: 0;
            padding: 20px;
            max-width: 1000px;
            margin: 0 auto;
        }
        h1, h2, h3 {
            color: #333;
        }
        .search-container {
            margin: 20px 0;
        }
        input[type="text"] {
            padding: 8px;
            width: 300px;
            font-size: 16px;
        }
        button {
            padding: 8px 16px;
            background-color: #4CAF50;
            color: white;
            border: none;
            cursor: pointer;
            font-size: 16px;
        }
        .result {
            margin-top: 20px;
            border: 1px solid #ddd;
            padding: 20px;
            border-radius: 5px;
        }
        .word-list {
            margin-top: 10px;
            max-height: 300px;
            overflow-y: auto;
            display: none;
        }
        .expanded .word-list {
            display: block;
            column-count: 3;
            column-gap: 20px;
        }
        .word-item {
            margin-bottom: 5px;
        }
        .word-item a {
            text-decoration: none;
            color: #0066cc;
        }
        .word-item a:hover {
            text-decoration: underline;
        }
        .stats {
            margin-bottom: 20px;
            color: #666;
        }
        .section-header {
            cursor: pointer;
            user-select: none;
            padding: 8px;
            background-color: #f5f5f5;
            border-radius: 4px;
        }
        .section-header:hover {
            background-color: #e5e5e5;
        }
        .section-header:after {
            content: " ▼";
            float: right;
            color: #666;
        }
        .expanded .section-header:after {
            content: " ▲";
        }
        .metrics {
            margin: 15px 0;
            padding: 15px;
            background-color: #f9f9f9;
            border-radius: 4px;
            border-left: 4px solid #4CAF50;
        }
        .metric-value {
            font-weight: bold;
        }
        .definitions {
            margin: 15px 0;
            column-count: 1 !important;
        }
        .definition {
            background-color: #f9f9f9;
            padding: 15px;
            margin-bottom: 15px;
            border-radius: 4px;
            border-left: 3px solid #0066cc;
            break-inside: avoid;
        }
        .definition-number {
            font-weight: bold;
            margin-right: 5px;
            color: #666;
        }
        .definition-words {
            margin-top: 10px;
            padding-top: 10px;
            border-top: 1px dashed #ccc;
            font-size: 0.9em;
        }
        .definition-words span {
            font-weight: bold;
            color: #666;
        }
        .definition-word {
            display: inline-block;
            margin: 2px 5px;
        }
        .definition-metrics {
            margin-top: 10px;
            padding-top: 10px;
            border-top: 1px dashed #ccc;
            font-size: 0.9em;
        }
        .metric-label {
            font-weight: bold;
            color: #666;
        }
        nav {
            background-color: #333;
            color: white;
            padding: 10px;
            margin-bottom: 20px;
            border-radius: 4px;
        }
        nav a {
            color: white;
            text-decoration: none;
            padding: 0 15px;
        }
        nav a:hover {
            text-decoration: underline;
        }
        table {
            width: 100%;
            border-collapse: collapse;
            margin: 20px 0;
        }
        th, td {
            padding: 10px;
            text-align: left;
            border-bottom: 1px solid #ddd;
        }
        th {
            background-color: #f5f5f5;
            font-weight: bold;
        }
        tr:hover {
            background-color: #f9f9f9;
        }
        .extreme {
            font-style: italic;
            color: #ff4500;
        }
        .tab-container {
            margin: 20px 0;
        }
        .tab-buttons {
            display: flex;
            gap: 10px;
            margin-bottom: 20px;
        }
        .tab-button {
            padding: 10px 20px;
            background-color: #f5f5f5;
            border: none;
            border-radius: 4px;
            cursor: pointer;
            font-size: 16px;
        }
        .tab-button.active {
            background-color: #4CAF50;
            color: white;
        }
        .tab-content {
            display: none;
        }
        .tab-content.active {
            display: block;
        }
        #graph-container {
            width: 100%;
            height: 600px;
            border: 1px solid #ddd;
            border-radius: 4px;
            overflow: hidden;
            position: relative;
        }
        .graph-controls {
            margin-bottom: 20px;
        }
        .node text {
            font-size: 12px;
            font-family: Arial, sans-serif;
        }
        .node circle {
            stroke: #fff;
            stroke-width: 2px;
        }
        .node.central circle {
            stroke: #4CAF50;
            stroke-width: 3px;
        }
        .node.second-hop circle {
            opacity: 0.6;
        }
        .link {
            stroke: #999;
            stroke-opacity: 0.6;
        }
        .link.second-hop {
            stroke-opacity: 0.3;
            stroke-dasharray: 3,3;
        }
        .link.side-connection {
            stroke: #ccc;
            stroke-opacity: 0.2;
        }
        .link.side-connection.second-hop {
            stroke-opacity: 0.1;
        }
        .graph-legend {
            position: absolute;
            top: 10px;
            right: 10px;
            background: rgba(255, 255, 255, 0.9);
            padding: 10px;
            border-radius: 4px;
            border: 1px solid #ddd;
        }
        .legend-item {
            margin: 5px 0;
            display: flex;
            align-items: center;
        }
        .legend-color {
            width: 20px;
            height: 20px;
            border-radius: 50%;
            margin-right: 10px;
        }
    </style>
    <script>
        function toggleSection(id) {
            const section = document.getElementById(id);
            section.classList.toggle('expanded');
        }

        function switchTab(tabName) {
            // Hide all tabs
            document.querySelectorAll('.tab-content').forEach(tab => {
                tab.classList.remove('active');
            });
            document.querySelectorAll('.tab-button').forEach(button => {
                button.classList.remove('active');
            });

            // Show selected tab
            document.getElementById(tabName + '-tab').classList.add('active');
            document.getElementById(tabName + '-button').classList.add('active');
        }

        function initGraph() {
            const container = document.getElementById('graph-container');
            const width = container.clientWidth;
            const height = container.clientHeight;
            
            // Clear any existing SVG
            container.innerHTML = '';
            
            const svg = d3.select('#graph-container')
                .append('svg')
                .attr('width', width)
                .attr('height', height);
            
            const g = svg.append('g');
            
            // Add zoom behavior
            const zoom = d3.zoom()
                .scaleExtent([0.1, 4])
                .on('zoom', (event) => {
                    g.attr('transform', event.transform);
                });
            
            svg.call(zoom);
            
            // Create the force simulation
            const simulation = d3.forceSimulation()
                .force('link', d3.forceLink().id(d => d.id).distance(100))
                .force('charge', d3.forceManyBody().strength(-300))
                .force('center', d3.forceCenter(width / 2, height / 2))
                .force('collision', d3.forceCollide().radius(30));
            
            function updateGraph(word) {
                // Fetch graph data
                fetch('/api/graph?word=' + encodeURIComponent(word))
                    .then(response => response.json())
                    .then(data => {
                        // Create the visualization
                        const link = g.selectAll('.link')
                            .data(data.links)
                            .join('line')
                            .attr('class', 'link');
                        
                        const node = g.selectAll('.node')
                            .data(data.nodes)
                            .join('g')
                            .attr('class', d => 'node ' + d.group)
                            .call(d3.drag()
                                .on('start', dragstarted)
                                .on('drag', dragged)
                                .on('end', dragended));
                        
                        node.append('circle')
                            .attr('r', d => Math.max(5, Math.min(20, Math.sqrt(d.size))))
                            .style('fill', d => {
                                if (d.group === 'central') return '#4CAF50';
                                if (d.group === 'defined') return '#2196F3';
                                if (d.group === 'defined-2hop') return '#90CAF9';
                                if (d.group === 'defines-2hop') return '#FFCC80';
                                return '#FFA726';
                            });
                        
                        node.append('text')
                            .attr('dx', 12)
                            .attr('dy', '.35em')
                            .text(d => d.label);
                        
                        // Update simulation
                        simulation
                            .nodes(data.nodes)
                            .on('tick', ticked);
                        
                        simulation.force('link')
                            .links(data.links);
                        
                        // Reset simulation
                        simulation.alpha(1).restart();
                        
                        function ticked() {
                            link
                                .attr('x1', d => d.source.x)
                                .attr('y1', d => d.source.y)
                                .attr('x2', d => d.target.x)
                                .attr('y2', d => d.target.y);
                            
                            node
                                .attr('transform', function(d) {
                                    return 'translate(' + d.x + ',' + d.y + ')';
                                });
                        }
                        
                        function dragstarted(event) {
                            if (!event.active) simulation.alphaTarget(0.3).restart();
                            event.subject.fx = event.subject.x;
                            event.subject.fy = event.subject.y;
                        }
                        
                        function dragged(event) {
                            event.subject.fx = event.x;
                            event.subject.fy = event.y;
                        }
                        
                        function dragended(event) {
                            if (!event.active) simulation.alphaTarget(0);
                            event.subject.fx = null;
                            event.subject.fy = null;
                        }

                        link.style('stroke-dasharray', d => d.isSecondHop ? '3,3' : null)
                            .style('stroke-opacity', d => {
                                if (d.isSideConnection && d.isSecondHop) return 0.1;
                                if (d.isSideConnection) return 0.2;
                                if (d.isSecondHop) return 0.3;
                                return 0.6;
                            })
                            .style('stroke', d => d.isSideConnection ? '#ccc' : '#999');
                    });
            }
            
            // Add event listener for the search form
            document.getElementById('graph-search-form').addEventListener('submit', function(e) {
                e.preventDefault();
                const word = document.getElementById('graph-search-input').value;
                updateGraph(word);
            });
        }
    </script>
</head>
<body>
    <nav>
        <a href="/">Home</a>
        <a href="/analysis">Analysis</a>
    </nav>
    
    {{if .IsHomePage}}
    <h1>Word Graph Explorer</h1>
    
    <div class="stats">
        <p>Graph contains {{.TotalLemmas}} lemmas with {{.TotalRelationships}} relationships</p>
    </div>
    
    <div class="search-container">
        <form action="/" method="GET">
            <input type="text" name="word" placeholder="Enter a word..." value="{{.Query}}">
            <button type="submit">Search</button>
        </form>
    </div>
    
    {{if .Lemma}}
    <div class="result">
        <h2>Information for '{{.Query}}' (lemma: '{{.Lemma}}')</h2>
        
        {{if gt .TotalWordRelations 0}}
        <div class="metrics">
            <h3>Highlights</h3>
            <p>Total relationships: <span class="metric-value">{{.TotalWordRelations}}</span></p>
            
            {{if eq .DefinesCount 0}}
            <p>This word only appears in definitions, never defines other words</p>
            {{else if eq .DefinedByCount 0}}
            <p>This word only defines other words, never appears in definitions</p>
            {{else}}
            <p>Ratio (defined-by:defines): <span class="metric-value">{{.Ratio}}:1</span></p>
            
            {{if gt .Ratio 2.0}}
            <p>This word is primarily used in definitions (consumer)</p>
            {{else if lt .Ratio 0.5}}
            <p>This word primarily defines others (producer)</p>
            {{else}}
            <p>This word has a balanced relationship in the lexicon</p>
            {{end}}
            {{end}}
        </div>
        {{end}}
        
        {{if gt (len .DefinitionData) 0}}
        <div id="definitions-section">
            <h3 class="section-header" onclick="toggleSection('definitions-section')">Definitions ({{len .DefinitionData}})</h3>
            <div class="word-list definitions">
                {{range $index, $def := .DefinitionData}}
                <div class="definition">
                    <div><span class="definition-number">{{inc $index}}.</span> {{$def.Text}}</div>
                    {{if gt $def.WordCount 0}}
                    <div class="definition-words">
                        <span>Words used:</span>
                        {{range $def.UsesWords}}
                        <a href="/?word={{.}}" class="definition-word">{{.}}</a>
                        {{end}}
                    </div>
                    <div class="definition-metrics">
                        <p><span class="metric-label">Words used:</span> <span class="metric-value">{{$def.WordCount}}</span></p>
                        <p><span class="metric-label">Definition ratio (in:out):</span> <span class="metric-value">{{$def.Ratio}}:1</span></p>
                        <p><span class="metric-label">Analysis:</span> {{$def.Analysis}}</p>
                    </div>
                    {{end}}
                </div>
                {{end}}
            </div>
        </div>
        {{end}}
        
        <div id="defined-by-section">
            <h3 class="section-header" onclick="toggleSection('defined-by-section')">Defined by ({{.DefinedByCount}} lemmas)</h3>
            <div class="word-list">
                {{range .DefinedByLemmas}}
                <div class="word-item"><a href="/?word={{.}}">{{.}}</a></div>
                {{end}}
            </div>
        </div>
        
        <div id="defines-section">
            <h3 class="section-header" onclick="toggleSection('defines-section')">Defines ({{.DefinesCount}} lemmas)</h3>
            <div class="word-list">
                {{range .DefinesLemmas}}
                <div class="word-item"><a href="/?word={{.}}">{{.}}</a></div>
                {{end}}
            </div>
        </div>
    </div>
    {{else if .Query}}
    <div class="result">
        <p>No information found for '{{.Query}}'</p>
    </div>
    {{end}}
    
    {{else if .IsAnalysisPage}}
    <h1>Word Graph Analysis</h1>
    
    <div class="stats">
        <p>Graph contains {{.TotalLemmas}} lemmas with {{.TotalRelationships}} relationships</p>
    </div>
    
    <div class="result">
        <form action="/analysis" method="GET">
            <label for="limit">Results to show:</label>
            <input type="number" name="limit" id="limit" value="{{.Limit}}" min="10" max="500" step="10">
            <label for="min_words">Minimum relationships:</label>
            <input type="number" name="min_words" id="min_words" value="{{.MinWords}}" min="1" max="100">
            <button type="submit">Update</button>
        </form>

        <div class="tab-container">
            <div class="tab-buttons">
                <button id="producers-button" class="tab-button active" onclick="switchTab('producers')">Producers</button>
                <button id="consumers-button" class="tab-button" onclick="switchTab('consumers')">Consumers</button>
                <button id="graph-button" class="tab-button" onclick="switchTab('graph'); initGraph();">Graph View</button>
            </div>

            <div id="producers-tab" class="tab-content active">
                <h2>Top Producers</h2>
                <p>These words primarily define other words but are rarely used in definitions themselves</p>
                <table>
                    <thead>
                        <tr>
                            <th>Lemma</th>
                            <th>Defined By</th>
                            <th>Defines</th>
                            <th>Ratio (in:out)</th>
                            <th>Total Relationships</th>
                        </tr>
                    </thead>
                    <tbody>
                        {{range .TopProducers}}
                        <tr>
                            <td><a href="/?word={{.Lemma}}">{{.Lemma}}</a></td>
                            <td>{{.DefinedBy}}</td>
                            <td>{{.Defines}}</td>
                            <td>{{if .IsExtreme}}<span class="extreme">pure producer</span>{{else}}{{printf "%.2f:1" .Ratio}}{{end}}</td>
                            <td>{{.WordCount}}</td>
                        </tr>
                        {{end}}
                    </tbody>
                </table>
            </div>

            <div id="consumers-tab" class="tab-content">
                <h2>Top Consumers</h2>
                <p>These words are frequently used in definitions but rarely define other words</p>
                <table>
                    <thead>
                        <tr>
                            <th>Lemma</th>
                            <th>Defined By</th>
                            <th>Defines</th>
                            <th>Ratio (in:out)</th>
                            <th>Total Relationships</th>
                        </tr>
                    </thead>
                    <tbody>
                        {{range .TopConsumers}}
                        <tr>
                            <td><a href="/?word={{.Lemma}}">{{.Lemma}}</a></td>
                            <td>{{.DefinedBy}}</td>
                            <td>{{.Defines}}</td>
                            <td>{{if .IsExtreme}}<span class="extreme">pure consumer</span>{{else}}{{printf "%.2f:1" .Ratio}}{{end}}</td>
                            <td>{{.WordCount}}</td>
                        </tr>
                        {{end}}
                    </tbody>
                </table>
            </div>

            <div id="graph-tab" class="tab-content">
                <h2>Graph Visualization</h2>
                <p>Explore the relationships between words in an interactive graph view</p>
                
                <div class="graph-controls">
                    <form id="graph-search-form">
                        <input type="text" id="graph-search-input" placeholder="Enter a word to visualize..." style="width: 300px;">
                        <button type="submit">Visualize</button>
                    </form>
                </div>
                
                <div id="graph-container">
                    <div class="graph-legend">
                        <div class="legend-item">
                            <div class="legend-color" style="background: #4CAF50;"></div>
                            <span>Selected Word</span>
                        </div>
                        <div class="legend-item">
                            <div class="legend-color" style="background: #2196F3;"></div>
                            <span>Defined by Selected</span>
                        </div>
                        <div class="legend-item">
                            <div class="legend-color" style="background: #90CAF9;"></div>
                            <span>Second-hop Defined</span>
                        </div>
                        <div class="legend-item">
                            <div class="legend-color" style="background: #FFA726;"></div>
                            <span>Defines Selected</span>
                        </div>
                        <div class="legend-item">
                            <div class="legend-color" style="background: #FFCC80;"></div>
                            <span>Second-hop Defines</span>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <div class="metrics">
            <h3>Additional Statistics</h3>
            <p>Words that only define others (pure producers): <span class="metric-value">{{.ExtremeProducers}}</span></p>
            <p>Words that are only defined by others (pure consumers): <span class="metric-value">{{.ExtremeConsumers}}</span></p>
        </div>
    </div>
    {{end}}
</body>
</html>
`))

// DefinitionData represents formatted definition data for the template
type DefinitionData struct {
	Text      string
	DefinesID string
	UsesWords []semantics.Lemma
	WordCount int
	Ratio     float64
	Analysis  string
}

// PageData represents the data for the home page template
type PageData struct {
	Query              string
	Lemma              semantics.Lemma
	DefinedByLemmas    []semantics.Lemma
	DefinesLemmas      []semantics.Lemma
	DefinitionData     []DefinitionData
	DefinedByCount     int
	DefinesCount       int
	TotalLemmas        int
	TotalRelationships int
	TotalWordRelations int
	Ratio              float64
	IsHomePage         bool
	IsAnalysisPage     bool

	// Analysis page data
	TopProducers     []LemmaRatio
	TopConsumers     []LemmaRatio
	Limit            int
	MinWords         int
	ExtremeProducers int
	ExtremeConsumers int
}

// analysisHandler serves the analysis page
func analysisHandler(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	minWordsStr := r.URL.Query().Get("min_words")

	limit := 100
	minWords := 5

	if limitStr != "" {
		if val, err := strconv.Atoi(limitStr); err == nil && val > 0 {
			limit = val
		}
	}

	if minWordsStr != "" {
		if val, err := strconv.Atoi(minWordsStr); err == nil && val > 0 {
			minWords = val
		}
	}

	// Get graph stats
	stats := wordGraph.GetStatistics()

	// Calculate ratios for all lemmas
	lemmaRatios := calculateLemmaRatios(wordGraph, minWords)

	// Count extreme cases
	extremeProducers := 0
	extremeConsumers := 0
	for _, lr := range lemmaRatios {
		if lr.IsExtreme {
			if lr.Category == "producer" {
				extremeProducers++
			} else if lr.Category == "consumer" {
				extremeConsumers++
			}
		}
	}

	// Sort for producers (first by ratio, then by number of words defined)
	producerRatios := make([]LemmaRatio, len(lemmaRatios))
	copy(producerRatios, lemmaRatios)
	sort.Slice(producerRatios, func(i, j int) bool {
		// First compare by ratio (lower is better for producers)
		if producerRatios[i].Ratio != producerRatios[j].Ratio {
			return producerRatios[i].Ratio < producerRatios[j].Ratio
		}
		// Within same ratio, sub-rank by number of words defined (higher is better)
		return producerRatios[i].Defines > producerRatios[j].Defines
	})

	// Sort for consumers (descending ratio)
	consumerRatios := make([]LemmaRatio, len(lemmaRatios))
	copy(consumerRatios, lemmaRatios)
	sort.Slice(consumerRatios, func(i, j int) bool {
		return consumerRatios[i].Ratio > consumerRatios[j].Ratio
	})

	// Limit results
	topProducers := producerRatios
	if len(topProducers) > limit {
		topProducers = topProducers[:limit]
	}

	topConsumers := consumerRatios
	if len(topConsumers) > limit {
		topConsumers = topConsumers[:limit]
	}

	data := PageData{
		TotalLemmas:        stats["TotalLemmas"],
		TotalRelationships: stats["TotalRelationships"],
		IsAnalysisPage:     true,
		TopProducers:       topProducers,
		TopConsumers:       topConsumers,
		Limit:              limit,
		MinWords:           minWords,
		ExtremeProducers:   extremeProducers,
		ExtremeConsumers:   extremeConsumers,
	}

	if err := templates.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// calculateLemmaRatios calculates ratio metrics for all lemmas
func calculateLemmaRatios(wordGraph *graph.WordGraph, minWords int) []LemmaRatio {
	lemmaRatios := make([]LemmaRatio, 0, len(wordGraph.Nodes))

	for _, node := range wordGraph.Nodes {
		definedByCount := len(node.DefinedByLemmas)
		definesCount := len(node.DefinesLemmas)
		totalCount := definedByCount + definesCount

		// Skip lemmas with too few relationships
		if totalCount < minWords {
			continue
		}

		// Calculate ratio and determine category
		ratio := 0.0
		category := "balanced"
		isExtreme := false

		if definesCount == 0 {
			ratio = float64(totalCount) // Max possible ratio
			category = "consumer"
			isExtreme = true
		} else if definedByCount == 0 {
			ratio = 0.0
			category = "producer"
			isExtreme = true
		} else {
			ratio = float64(definedByCount) / float64(definesCount)
			if ratio > 2.0 {
				category = "consumer"
			} else if ratio < 0.5 {
				category = "producer"
			}
		}

		lemmaRatios = append(lemmaRatios, LemmaRatio{
			Lemma:     node.Lemma,
			DefinedBy: definedByCount,
			Defines:   definesCount,
			Ratio:     ratio,
			WordCount: totalCount,
			Category:  category,
			IsExtreme: isExtreme,
		})
	}

	return lemmaRatios
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("word")

	// Get graph stats
	stats := wordGraph.GetStatistics()

	data := PageData{
		Query:              query,
		TotalLemmas:        stats["TotalLemmas"],
		TotalRelationships: stats["TotalRelationships"],
		IsHomePage:         true,
	}

	if query != "" {
		lemma := semantics.Lemmatize(query)
		node := wordGraph.GetLemmaNode(lemma)

		if node != nil && (len(node.DefinedByLemmas) > 0 || len(node.DefinesLemmas) > 0 || len(node.Definitions) > 0) {
			data.Lemma = lemma

			// Format definitions for display
			data.DefinitionData = make([]DefinitionData, 0, len(node.Definitions))
			for _, def := range node.Definitions {
				// Extract used words and sort them
				usesWords := make([]semantics.Lemma, 0, len(def.UsesWords))
				for word := range def.UsesWords {
					usesWords = append(usesWords, word)
				}
				sort.Slice(usesWords, func(i, j int) bool {
					return strings.ToLower(string(usesWords[i])) < strings.ToLower(string(usesWords[j]))
				})

				// Calculate ratio and analysis for this definition
				wordCount := len(def.UsesWords)
				ratio := 0.0
				analysis := ""

				if wordCount > 0 {
					ratio = float64(len(node.DefinedByLemmas)) / float64(wordCount)

					// Round to 2 decimal places
					ratio = float64(int(ratio*100)) / 100

					// Provide semantic analysis
					if ratio > 2.0 {
						analysis = "This definition is primarily using common defining words (abstract)"
					} else if ratio < 0.5 {
						analysis = "This definition uses many specific words (concrete)"
					} else {
						analysis = "This definition has a balanced vocabulary"
					}
				}

				defData := DefinitionData{
					Text:      def.Text,
					DefinesID: def.DefinesID,
					UsesWords: usesWords,
					WordCount: wordCount,
					Ratio:     ratio,
					Analysis:  analysis,
				}
				data.DefinitionData = append(data.DefinitionData, defData)
			}

			// Get and sort the words that define this lemma
			definedBy := make([]semantics.Lemma, 0, len(node.DefinedByLemmas))
			for l := range node.DefinedByLemmas {
				definedBy = append(definedBy, l)
			}
			sort.Slice(definedBy, func(i, j int) bool {
				return strings.ToLower(string(definedBy[i])) < strings.ToLower(string(definedBy[j]))
			})
			data.DefinedByLemmas = definedBy
			data.DefinedByCount = len(definedBy)

			// Get and sort the words this lemma defines
			defines := make([]semantics.Lemma, 0, len(node.DefinesLemmas))
			for l := range node.DefinesLemmas {
				defines = append(defines, l)
			}
			sort.Slice(defines, func(i, j int) bool {
				return strings.ToLower(string(defines[i])) < strings.ToLower(string(defines[j]))
			})
			data.DefinesLemmas = defines
			data.DefinesCount = len(defines)

			// Calculate metrics similar to the lookup command
			data.TotalWordRelations = data.DefinedByCount + data.DefinesCount

			// Calculate ratio and handle division by zero
			if data.DefinesCount > 0 {
				data.Ratio = float64(data.DefinedByCount) / float64(data.DefinesCount)
				// Format to 2 decimal places
				data.Ratio = float64(int(data.Ratio*100)) / 100
			}
		}
	}

	if err := templates.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	// Parse command line flags
	graphFile := flag.String("graph", "./data/wordgraph.gz", "Path to the compressed graph file")
	port := flag.Int("port", 8080, "Port to run the web server on")
	flag.Parse()

	// Load the word graph
	var err error
	wordGraph, err = graph.LoadCompressed(*graphFile)
	if err != nil {
		log.Fatalf("Error loading graph: %v", err)
	}

	// Set up the HTTP server
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/analysis", analysisHandler)
	http.HandleFunc("/api/graph", graphDataHandler)

	// Start the server
	addr := fmt.Sprintf(":%d", *port)
	fmt.Printf("Starting web UI on http://localhost%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

// GraphNode represents a node in the graph visualization
type GraphNode struct {
	ID    string  `json:"id"`
	Label string  `json:"label"`
	Group string  `json:"group"`
	Size  int     `json:"size"`
	Ratio float64 `json:"ratio"`
}

// GraphLink represents a link in the graph visualization
type GraphLink struct {
	Source           string `json:"source"`
	Target           string `json:"target"`
	IsSecondHop      bool   `json:"isSecondHop"`
	IsSideConnection bool   `json:"isSideConnection"`
}

// GraphData represents the complete graph data for visualization
type GraphData struct {
	Nodes []GraphNode `json:"nodes"`
	Links []GraphLink `json:"links"`
}

// graphDataHandler serves graph data for visualization
func graphDataHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("word")
	limit := 25 // Reduced limit per hop to prevent overload

	var graphData GraphData

	if query != "" {
		// Get the central node
		lemma := semantics.Lemmatize(query)
		node := wordGraph.GetLemmaNode(lemma)

		if node != nil {
			addedNodes := make(map[string]bool)
			firstHopNodes := make(map[string]*semantics.LemmaNode) // Changed to store node pointers

			// Add the central node
			ratio := 0.0
			if len(node.DefinesLemmas) > 0 {
				ratio = float64(len(node.DefinedByLemmas)) / float64(len(node.DefinesLemmas))
			}

			graphData.Nodes = append(graphData.Nodes, GraphNode{
				ID:    string(node.Lemma),
				Label: string(node.Lemma),
				Group: "central",
				Size:  len(node.DefinedByLemmas) + len(node.DefinesLemmas),
				Ratio: ratio,
			})
			addedNodes[string(node.Lemma)] = true

			// First hop: Add nodes that this word defines
			nodeCount := 0
			for lemma := range node.DefinesLemmas {
				if nodeCount >= limit {
					break
				}
				targetNode := wordGraph.GetLemmaNode(lemma)
				if targetNode != nil {
					if !addedNodes[string(lemma)] {
						ratio := 0.0
						if len(targetNode.DefinesLemmas) > 0 {
							ratio = float64(len(targetNode.DefinedByLemmas)) / float64(len(targetNode.DefinesLemmas))
						}
						graphData.Nodes = append(graphData.Nodes, GraphNode{
							ID:    string(lemma),
							Label: string(lemma),
							Group: "defined",
							Size:  len(targetNode.DefinedByLemmas) + len(targetNode.DefinesLemmas),
							Ratio: ratio,
						})
						addedNodes[string(lemma)] = true
						nodeCount++
					}
					firstHopNodes[string(lemma)] = targetNode
					graphData.Links = append(graphData.Links, GraphLink{
						Source: string(node.Lemma),
						Target: string(lemma),
					})
				}
			}

			// First hop: Add nodes that define this word
			nodeCount = 0
			for lemma := range node.DefinedByLemmas {
				if nodeCount >= limit {
					break
				}
				sourceNode := wordGraph.GetLemmaNode(lemma)
				if sourceNode != nil {
					if !addedNodes[string(lemma)] {
						ratio := 0.0
						if len(sourceNode.DefinesLemmas) > 0 {
							ratio = float64(len(sourceNode.DefinedByLemmas)) / float64(len(sourceNode.DefinesLemmas))
						}
						graphData.Nodes = append(graphData.Nodes, GraphNode{
							ID:    string(lemma),
							Label: string(lemma),
							Group: "defines",
							Size:  len(sourceNode.DefinedByLemmas) + len(sourceNode.DefinesLemmas),
							Ratio: ratio,
						})
						addedNodes[string(lemma)] = true
						nodeCount++
					}
					firstHopNodes[string(lemma)] = sourceNode
					graphData.Links = append(graphData.Links, GraphLink{
						Source: string(lemma),
						Target: string(node.Lemma),
					})
				}
			}

			// Add connections between first-hop nodes
			for lemmaA, nodeA := range firstHopNodes {
				for lemmaB, _ := range firstHopNodes {
					if lemmaA != lemmaB {
						// Check if A defines B
						if _, ok := nodeA.DefinesLemmas[semantics.Lemma(lemmaB)]; ok {
							graphData.Links = append(graphData.Links, GraphLink{
								Source:           lemmaA,
								Target:           lemmaB,
								IsSecondHop:      true,
								IsSideConnection: true,
							})
						}
					}
				}
			}

			// Second hop: Add connections from first hop nodes
			secondHopNodes := make(map[string]*semantics.LemmaNode)
			for firstHopLemma, firstHopNode := range firstHopNodes {
				// Add nodes that the first hop node defines
				nodeCount = 0
				for lemma := range firstHopNode.DefinesLemmas {
					if nodeCount >= limit/2 {
						break
					}
					targetNode := wordGraph.GetLemmaNode(lemma)
					if targetNode != nil && !addedNodes[string(lemma)] {
						ratio := 0.0
						if len(targetNode.DefinesLemmas) > 0 {
							ratio = float64(len(targetNode.DefinedByLemmas)) / float64(len(targetNode.DefinesLemmas))
						}
						graphData.Nodes = append(graphData.Nodes, GraphNode{
							ID:    string(lemma),
							Label: string(lemma),
							Group: "defined-2hop",
							Size:  len(targetNode.DefinedByLemmas) + len(targetNode.DefinesLemmas),
							Ratio: ratio,
						})
						graphData.Links = append(graphData.Links, GraphLink{
							Source:      string(firstHopLemma),
							Target:      string(lemma),
							IsSecondHop: true,
						})
						addedNodes[string(lemma)] = true
						secondHopNodes[string(lemma)] = targetNode
						nodeCount++
					} else if targetNode != nil && addedNodes[string(lemma)] {
						// Add connection if target exists in graph
						graphData.Links = append(graphData.Links, GraphLink{
							Source:      string(firstHopLemma),
							Target:      string(lemma),
							IsSecondHop: true,
						})
					}
				}

				// Add nodes that define the first hop node
				nodeCount = 0
				for lemma := range firstHopNode.DefinedByLemmas {
					if nodeCount >= limit/2 {
						break
					}
					sourceNode := wordGraph.GetLemmaNode(lemma)
					if sourceNode != nil && !addedNodes[string(lemma)] {
						ratio := 0.0
						if len(sourceNode.DefinesLemmas) > 0 {
							ratio = float64(len(sourceNode.DefinedByLemmas)) / float64(len(sourceNode.DefinesLemmas))
						}
						graphData.Nodes = append(graphData.Nodes, GraphNode{
							ID:    string(lemma),
							Label: string(lemma),
							Group: "defines-2hop",
							Size:  len(sourceNode.DefinedByLemmas) + len(sourceNode.DefinesLemmas),
							Ratio: ratio,
						})
						graphData.Links = append(graphData.Links, GraphLink{
							Source:      string(lemma),
							Target:      string(firstHopLemma),
							IsSecondHop: true,
						})
						addedNodes[string(lemma)] = true
						secondHopNodes[string(lemma)] = sourceNode
						nodeCount++
					} else if sourceNode != nil && addedNodes[string(lemma)] {
						// Add connection if source exists in graph
						graphData.Links = append(graphData.Links, GraphLink{
							Source:      string(lemma),
							Target:      string(firstHopLemma),
							IsSecondHop: true,
						})
					}
				}
			}

			// Add connections between second-hop nodes and all other nodes
			for lemmaA, nodeA := range secondHopNodes {
				// Check connections to first-hop nodes
				for lemmaB, _ := range firstHopNodes {
					if lemmaA != lemmaB {
						// Check if A defines B
						if _, ok := nodeA.DefinesLemmas[semantics.Lemma(lemmaB)]; ok {
							graphData.Links = append(graphData.Links, GraphLink{
								Source:           lemmaA,
								Target:           lemmaB,
								IsSecondHop:      true,
								IsSideConnection: true,
							})
						}
					}
				}

				// Check connections to other second-hop nodes
				for lemmaB, _ := range secondHopNodes {
					if lemmaA != lemmaB {
						// Check if A defines B
						if _, ok := nodeA.DefinesLemmas[semantics.Lemma(lemmaB)]; ok {
							graphData.Links = append(graphData.Links, GraphLink{
								Source:           lemmaA,
								Target:           lemmaB,
								IsSecondHop:      true,
								IsSideConnection: true,
							})
						}
					}
				}
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(graphData)
}
