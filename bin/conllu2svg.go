package main

import (
	"bufio"
	"bytes"
	"fmt"
	"html"
	"os"
	"sort"
	"strings"
)

const (
	FONT = "FreeSans, Arial, Helvetica, sans-serif"

	MIN_NODE_WIDTH = 80 // minimale breedte van nodes
	NODE_HEIGHT    = 48 // hoogte van nodes
	NODE_SPACING   = 8  // horizontale ruimte tussen nodes
	NODE_FONT_SIZE = 16 // fontsize in nodes
	NODE_TWEEK     = 2  // schuif teksten verticaal naar elkaar toe

	LVL_HEIGHT             = 40      // hoogteverschil tussen edges van opeenvolgend niveau
	EDGE_FONT_SIZE         = 16      // fontsize van label bij edge
	EDGE_FONT_OFFSET       = 8       // hoogte van baseline van label boven edge
	EDGE_FONT_WHITE_MARGIN = 2       // extra witruimte om label bij edge
	EDGE_LBL_BACKGROUND    = "white" // kleur van rechthoek achter label boven edge
	EDGE_LBL_OPACITY       = .9      // doorzichtigheid van rechthoek achter label boven edge
	EDGE_DROP              = 80      // edge curvature: te veel, en lijnen steken onder de figuur uit

	MULTI_SKIP   = 4
	MULTI_HEIGHT = 28

	MARGIN = 4 // extra ruimte rond hele figuur

	TESTING = false
)

type Item struct {
	lineno   int
	here     string
	there    string
	end      int
	enhanced bool
	word     string
	lemma    string
	postag   string
	xpostag  string
	attribs  string
	rel      string
	deps     string
	x1, x2   int
}

type Dependency struct {
	end     int
	headpos int
	rel     [2]string
	dist    int
	lvl     int
}

type Anchor struct {
	dist  int
	level int
}

type Line struct {
	text   string
	lineno int
}

type Multi struct {
	id     string
	word   string
	lineno int
}

var (
	dependencies []*Dependency
	anchors      [][]Anchor
	columns      = make(map[string]int)
	nColumns     int
)

func main() {
	fp := os.Stdout
	fmt.Fprint(fp, `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<meta name="robots" content="noindex,nofollow">
<title>Universal Dependencies</title>
<link rel="stylesheet" type="text/css" href="../conllu/tooltip.css" />
<script type="text/javascript" src="../conllu/jquery.js"></script>
<script type="text/javascript" src="../conllu/tooltip.js"></script>
<script type="text/javascript">
var tts = [];
var normal = [];
function toggle(id, enhanced) {
  unhold();
  if (normal[id] && enhanced) {
    $('svg#'+id+' .enhanced').css({'visibility':''});
    $('svg#'+id+' .normal').css({'visibility':'hidden'});
    normal[id] = false;
  } else if (!normal[id] && !enhanced) {
    $('svg#'+id+' .enhanced').css({'visibility':'hidden'});
    $('svg#'+id+' .normal').css({'visibility':''});
    normal[id] = true;
  }
}

var holding  = false;
var holdedge = false;
var holdnode = false;
var holdid  = '';
var holdi = -1;
var holdj = -1;

function unhold() {
  if (holding) {
    holding = false;
    if (holdnode) {
      holdnode = false;
      unmark(holdid, holdi);
    }
    if (holdedge) {
      holdedge = false;
      unmrk(holdid, holdi, holdj);
    }
    holdid = '';
    holdi = -1;
    holdj = -1;
  }
}
function hold(id, i) {
  if (holdnode && id == holdid && i == holdi) {
    unhold();
  } else {
    unhold();
    mark(id, i);
    holding = true;
    holdnode = true;
    holdid = id;
    holdi = i;
  }
}
function hld(id, i, j) {
  if (holdedge && id == holdid && i == holdi && j == holdj) {
    unhold();
  } else {
    unhold();
    mrk(id, i, j);
    holding = true;
    holdedge = true;
    holdid = id;
    holdi = i;
    holdj = j;
  }
}
function mark(id, i) {
  var cl = normal[id] ? 'n' : 'e';
  var t = tts[id][i-1];
  tooltip.show('<em>' + t[0] + '</em><br>\n' + t[1] + '<br>\n' + t[2] + '<br>\nLemma: ' + t[3] + (t[4] == "_" ? "" : '<br>\nXpos: ' + t[4]));
  if (!holding) {
    $('svg#' + id + ' .l' + cl + i).css({'fill':'blue','font-weight':'bold'});
    $('svg#' + id + ' .e' + cl + i).css({'stroke':'blue','stroke-width':3});
  }
}
function unmark(id, i) {
  var cl = normal[id] ? 'n' : 'e';
  tooltip.hide();
  if (!holding) {
    $('svg#' + id + ' .l' + cl + i).css({'fill':'black','font-weight':'normal'});
    $('svg#' + id + ' .e' + cl + i).css({'stroke':'black','stroke-width':1});
  }
}
function mrk(id, i, j) {
  if (!holding) {
    $('svg#' + id + ' .p' + i + 'p' + j).css({'fill':'blue','font-weight':'bold'});
    $('svg#' + id + ' .q' + i + 'q' + j).css({'stroke':'blue','stroke-width':3});
    $('svg#' + id + ' .q' + i).css({'stroke':'blue','stroke-width':3});
    $('svg#' + id + ' .q' + j).css({'stroke':'blue','stroke-width':3});
  }
}
function unmrk(id, i, j) {
  if (!holding) {
    $('svg#' + id + ' .p' + i + 'p' + j).css({'fill':'black','font-weight':'normal'});
    $('svg#' + id + ' .q' + i + 'q' + j).css({'stroke':'black','stroke-width':1});
    $('svg#' + id + ' .q' + i).css({'stroke':'black','stroke-width':1});
    $('svg#' + id + ' .q' + j).css({'stroke':'black','stroke-width':1});
  }
}
</script>
<style type="text/css">
  body { padding-bottom: 200px; }
  div.comments {
    padding: 1em;
    margin: 2em 0px 1em 0px;
    background-color: #f0f0ff;
    white-space: pre-line;
  }
  div.svg {
    margin: 40px 0px;
  }
  div.overflow {
    overflow-x: auto;
  }
  .udcontrol input,
  .udcontrol label {
    cursor: pointer;
  }
  .udcontrol label:hover {
    color: #0000e0;
    text-decoration: underline;
  }
</style>
</head>
<body>
`)

	setColumns("# global.columns = ID FORM LEMMA UPOS XPOS FEATS HEAD DEPREL DEPS MISC")
	id := 0
	for _, filename := range os.Args[1:] {
		if len(os.Args) > 2 {
			fmt.Printf("<h2>%s</h2>\n", html.EscapeString(filename))
		}
		linecount := 0
		var lines []Line
		inComments := false
		inDep := false
		fp, err := os.Open(filename)
		x(err, filename, 0)
		scanner := bufio.NewScanner(fp)
		for scanner.Scan() {
			linecount++
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				if inDep {
					id++
					doSentence(lines, filename, id)
					inDep = false
				}
				if inComments {
					fmt.Println("</div>")
					inComments = false
				}
			} else if line[0] == '#' {
				if f := strings.Fields(line); len(f) > 2 && f[1] == "global.columns" && f[2] == "=" {
					x(setColumns(line), filename, linecount)
				}
				if inDep {
					x(fmt.Errorf("Comment line in dependency"), filename, linecount)
				}
				if !inComments {
					fmt.Print(`<div class="comments">`)
					inComments = true
				}
				fmt.Println(html.EscapeString(line))
			} else {
				if inComments {
					fmt.Println("</div>")
					inComments = false
				}
				if !inDep {
					lines = make([]Line, 0)
					inDep = true
				}
				lines = append(lines, Line{text: line, lineno: linecount})
			}
		}
		x(scanner.Err(), filename, linecount)
		fp.Close()
		if inComments {
			fmt.Println("</div>")
			inComments = false
		}
		if inDep {
			x(fmt.Errorf("Missing blank line at end of dependency"), filename, linecount)
		}
	}

	fmt.Fprint(fp, `</body>
</html>
`)
	fp.Close()
}

func setColumns(s string) error {
	ff := strings.Fields(s)[3:]
	nColumns = len(ff)
	for _, c := range []string{"ID", "FORM", "LEMMA", "UPOS", "XPOS", "FEATS", "HEAD", "DEPREL", "DEPS", "MISC"} {
		columns[c] = nColumns
	}
	for i, c := range ff {
		columns[c] = i
	}
	for _, c := range []string{"ID", "FORM", "HEAD", "DEPREL"} {
		if columns[c] == nColumns {
			return fmt.Errorf("Missing required field %s in global.columns", c)
		}
	}
	return nil
}

func doSentence(lines []Line, filename string, id int) {
	dependencies = make([]*Dependency, 0)

	hasEnhanced := false
	svgID := fmt.Sprintf("svg%d", id)
	items := make([]*Item, 0)
	positions := map[string]int{
		"0": 0,
	}
	multis := make([]Multi, 0)

	n := 0
	for _, line := range lines {
		aa := strings.Split(line.text, "\t")
		if len(aa) < 2 {
			aa = strings.Fields(line.text)
		}
		if len(aa) != nColumns {
			x(fmt.Errorf("Wrong number of fields"), filename, line.lineno)
		}
		aa = append(aa, "_")
		for i, a := range aa {
			aa[i] = strings.TrimSpace(a)
		}

		if strings.Contains(aa[columns["ID"]], ".") && columns["DEPS"] == nColumns {
			continue
		}

		if strings.Contains(aa[columns["ID"]], "-") {
			multis = append(multis, Multi{id: aa[columns["ID"]], word: aa[columns["FORM"]], lineno: line.lineno})
			continue
		}
		at := ""
		if aa[columns["FEATS"]] != "_" {
			at = strings.Replace(strings.Replace(aa[columns["FEATS"]], "|", ", ", -1), "=", ": ", -1)
		}
		items = append(items, &Item{
			lineno:   line.lineno,
			here:     aa[columns["ID"]],
			word:     aa[columns["FORM"]],
			lemma:    aa[columns["LEMMA"]],
			postag:   aa[columns["UPOS"]],
			xpostag:  aa[columns["XPOS"]],
			attribs:  at,
			there:    aa[columns["HEAD"]],
			rel:      aa[columns["DEPREL"]],
			deps:     aa[columns["DEPS"]],
			enhanced: strings.Contains(aa[columns["ID"]], "."),
		})
		n++
		positions[aa[columns["ID"]]] = n
	}

	for i, item := range items {
		end := positions[item.here]
		items[i].end = end

		if !item.enhanced {
			headpos, ok := positions[item.there]
			if !ok {
				x(fmt.Errorf("Unknown head position %s", item.there), filename, item.lineno)
			}
			dependencies = append(dependencies, &Dependency{
				end:     end,
				headpos: headpos,
				rel:     [2]string{item.rel, ""},
				dist:    abs(end - headpos)})
		}

		if item.deps != "_" {
			parts := strings.Split(item.deps, "|")
			for _, part := range parts {
				a := strings.SplitN(part, ":", 2)
				if len(a) != 2 {
					x(fmt.Errorf("Invalid dependency: %s", part), filename, item.lineno)
				}
				headpos, ok := positions[a[columns["ID"]]]
				if !ok {
					x(fmt.Errorf("Unknown head position %s", a[columns["ID"]]), filename, item.lineno)
				}
				dependencies = append(dependencies, &Dependency{
					end:     end,
					headpos: headpos,
					rel:     [2]string{"", a[columns["FORM"]]},
					dist:    abs(end - headpos)})
				hasEnhanced = true
			}
		}
	}

	// dubbele edges samenvoegen
	for i := 0; i < len(dependencies); i++ {
		d1 := dependencies[i]
		if d1.rel[0] != "" {
			for j := 0; j < len(dependencies); j++ {
				if i == j {
					continue
				}
				d2 := dependencies[j]
				if d2.rel[1] != "" && d1.end == d2.end && d1.headpos == d2.headpos && d1.dist == d2.dist {
					d1.rel[1] = d2.rel[1]
					dependencies = append(dependencies[:j], dependencies[j+1:]...)
					if j < i {
						i--
					}
					break
				}
			}
		}
	}

	// posities van de nodes

	sort.Slice(items, func(i, j int) bool { return items[i].end < items[j].end })
	width := MARGIN
	for i, item := range items {
		if item.end != i+1 {
			x(fmt.Errorf("Wrong index: %d != %d", item.end, i+1), filename, item.lineno)
		}
		item.x1 = width
		w1, _, _ := textwidth(item.postag+" i", NODE_FONT_SIZE, false)
		w2, _, _ := textwidth(item.word+" i", NODE_FONT_SIZE, false)
		item.x2 = width + 24 + max(MIN_NODE_WIDTH, w1, w2)
		width = item.x2 + NODE_SPACING
	}
	width -= NODE_SPACING
	width += MARGIN

	// hoogtes van de edges en aangrijppunten van de edges

	anchors = make([][]Anchor, len(items))
	for i := range items {
		anchors[i] = make([]Anchor, 0)
	}

	sort.Slice(dependencies, func(i, j int) bool { return dependencies[i].dist < dependencies[j].dist })
	grid := make([][]bool, len(items))
	for i := range grid {
		grid[i] = make([]bool, 2*len(items))
	}
	for i, dep := range dependencies {
		if dep.headpos == 0 {
			anchors[dep.end-1] = append(anchors[dep.end-1], Anchor{})
			continue
		}
		i1, i2 := dep.end-1, dep.headpos-1
		if i1 > i2 {
			i1, i2 = i2, i1
		}
		lvl := 0
		for {
			ok := true
			for i := i1; i < i2; i++ {
				if grid[i][lvl] {
					ok = false
					break
				}
			}
			if ok {
				for i := i1; i < i2; i++ {
					grid[i][lvl] = true
				}
				break
			}
			lvl++
		}
		dependencies[i].lvl = lvl
		anchors[dep.end-1] = append(anchors[dep.end-1], Anchor{dist: dep.headpos - dep.end, level: lvl})
		anchors[dep.headpos-1] = append(anchors[dep.headpos-1], Anchor{dist: dep.end - dep.headpos, level: lvl})
	}

	maxlvl := 0
	for _, dep := range dependencies {
		maxlvl = max(maxlvl, dep.lvl)
	}
	if hasEnhanced {
		maxlvl++
	}

	// correctie voor root-dependencies
	for i, dep := range dependencies {
		if dep.headpos == 0 {
			dependencies[i].lvl = maxlvl
		}
	}
	for key, anchor := range anchors {
		for i, a := range anchor {
			if a.dist == 0 {
				anchors[key][i].level = maxlvl
			}
		}
	}

	for n := range anchors {
		sort.Slice(anchors[n], func(i, j int) bool {
			var a1 = anchors[n][i]
			var a2 = anchors[n][j]
			if a1.dist == 0 {
				return a2.dist > 0
			}
			if a2.dist == 0 {
				return a1.dist < 0
			}
			if a1.dist == a2.dist {
				if a1.dist < 0 {
					return a1.level < a2.level
				}
				return a1.level > a2.level
			}
			if a1.dist < 0 {
				if a2.dist > 0 {
					return true
				}
				if a1.dist < a2.dist {
					return false
				}
				return true
			}
			if a2.dist < 0 {
				return false
			}
			if a1.dist < a2.dist {
				return false
			}
			return true
		})
	}

	height := MARGIN + EDGE_FONT_SIZE + EDGE_FONT_OFFSET + LVL_HEIGHT*(maxlvl+1) + NODE_HEIGHT + MARGIN
	if len(multis) > 0 {
		height += MULTI_HEIGHT + MULTI_SKIP
	}

	// begin uitvoer

	fmt.Printf("<div class=\"overflow\">\n<svg id=\"%s\" width=\"%d\" height=\"%d\">\n", svgID, width, height)

	if TESTING {
		fmt.Printf("<rect x=\"0\" y=\"0\" width=\"%d\" height=\"%d\" fill=\"green\" />\n", width, height)
	}

	// edges

	for variant := 0; variant < 2; variant++ {

		e := "n"
		if variant == 1 {
			if !hasEnhanced {
				continue
			}
			e = "e"
		}

		var lines bytes.Buffer
		var arrows bytes.Buffer
		var whites bytes.Buffer
		var texts bytes.Buffer

		for _, dep := range dependencies {
			if dep.rel[variant] == "" {
				continue
			}
			i1, i2 := dep.end-1, dep.headpos-1
			if dep.headpos == 0 {
				i2 = i1
			}
			d1 := float64(items[i1].x2-items[i1].x1) - 20
			x1 := items[i1].x1 + 10 + int(d1*anchor(i1, i2, dep.lvl))
			d2 := float64(items[i2].x2-items[i2].x1) - 20
			x2 := items[i2].x1 + 10 + int(d2*anchor(i2, i1, dep.lvl))
			y1 := MARGIN + EDGE_FONT_SIZE + EDGE_FONT_OFFSET + LVL_HEIGHT*(maxlvl+1)
			y2 := MARGIN + EDGE_FONT_SIZE + EDGE_FONT_OFFSET + LVL_HEIGHT*(maxlvl-dep.lvl)
			if dep.headpos == 0 {
				y2 = MARGIN + EDGE_FONT_SIZE + EDGE_FONT_OFFSET
				fmt.Fprintf(&lines,
					"<path class=\"e%s%d q%dq%d\" d=\"M%d %d L%d %d\" />\n",
					e,
					dep.end,
					dep.end,
					dep.headpos,
					x1, y1, // M
					x1, y2) // L
			} else {
				fmt.Fprintf(&lines,
					"<path class=\"e%s%d e%s%d q%dq%d\" d=\"M%d %d L%d %d C%d %d %d %d %d %d C%d %d %d %d %d %d L%d %d\" />\n",
					e,
					dep.end,
					e,
					dep.headpos,
					dep.end,
					dep.headpos,
					x1, y1, // M
					x1, y2+EDGE_DROP, // L
					x1, y2, // C
					x1, y2,
					(x1+x2)/2, y2,
					x2, y2, // C
					x2, y2,
					x2, y2+EDGE_DROP,
					x2, y1) // L
			}
			fmt.Fprintf(&arrows,
				"<path class=\"e%s%d e%s%d q%dq%d\" d=\"M%d %d l3 -14 l-6 0 Z\" />\n",
				e,
				dep.end,
				e,
				dep.headpos,
				dep.end,
				dep.headpos,
				x1, y1)
			w, h, l := textwidth(dep.rel[variant]+"i", EDGE_FONT_SIZE, true)
			fmt.Fprintf(&whites,
				"<rect x=\"%d\" y=\"%d\" width=\"%d\" height=\"%d\" />\n",
				(x1+x2-w)/2-EDGE_FONT_WHITE_MARGIN,
				y2-l-EDGE_FONT_OFFSET-EDGE_FONT_WHITE_MARGIN,
				w+2*EDGE_FONT_WHITE_MARGIN,
				h+2*EDGE_FONT_WHITE_MARGIN)
			fmt.Fprintf(&texts,
				"<text class=\"l%s%d l%s%d p%dp%d\" x=\"%d\" y=\"%d\" "+
					"onclick=\"hld('%s',%d,%d)\" onmouseover=\"mrk('%s',%d,%d)\" onmouseout=\"unmrk('%s',%d,%d)\">"+
					"%s</text>\n",
				e,
				dep.end,
				e,
				dep.headpos,
				dep.end,
				dep.headpos,
				(x1+x2)/2,
				y2-EDGE_FONT_OFFSET,
				svgID, dep.end, dep.headpos,
				svgID, dep.end, dep.headpos,
				svgID, dep.end, dep.headpos,
				html.EscapeString(dep.rel[variant]))
		}

		if variant == 0 {
			fmt.Println("<g class=\"normal\" style=\"visibility:hidden\">")
		} else {
			fmt.Println("<g class=\"enhanced\">")
		}

		fmt.Print("<g fill=\"none\" stroke=\"black\" stroke-width=\"1\">\n", lines.String(), "</g>\n")
		fmt.Print("<g fill=\"black\" stroke-width=\"1\" stroke=\"black\">\n", arrows.String(), "</g>\n")
		fmt.Printf("<g fill=\"%s\" stroke=\"%s\" stroke-width=\"1\" opacity=\"%g\">\n%s</g>\n",
			EDGE_LBL_BACKGROUND,
			EDGE_LBL_BACKGROUND,
			EDGE_LBL_OPACITY,
			whites.String())
		fmt.Printf("<g font-family=\"%s\" font-size=\"%d\" text-anchor=\"middle\">\n%s</g>\n", FONT, int(EDGE_FONT_SIZE), texts.String())

		fmt.Println("</g>")
	}

	// nodes

	offset := MARGIN + EDGE_FONT_SIZE + EDGE_FONT_OFFSET + LVL_HEIGHT*(maxlvl+1)

	fmt.Println("<g fill=\"#D0E0FF\" stroke=\"black\" stroke-width=\"1\">")
	for i, item := range items {
		enh := ""
		color := ""
		if item.enhanced {
			enh = "enhanced "
			color = `fill="#ffe0e0" `
		}
		fmt.Printf("<rect class=\"%sq%d %s\" x=\"%d\" y=\"%d\" width=\"%d\" height=\"%d\" rx=\"5\" ry=\"5\" %s/>\n",
			enh,
			i+1,
			classlbl(item),
			item.x1,
			offset,
			item.x2-item.x1,
			int(NODE_HEIGHT),
			color)
	}
	fmt.Println("</g>")

	_, _, y := textwidth("Xg", NODE_FONT_SIZE, false)
	lower := y / 2

	fmt.Printf("<g font-family=\"%s\" font-size=\"%d\" text-anchor=\"middle\">\n", FONT, int(NODE_FONT_SIZE))
	for _, item := range items {
		enh := ""
		if item.enhanced {
			enh = ` class="enhanced"`
		}
		fmt.Printf("<text%s x=\"%d\" y=\"%d\">%s</text>\n",
			enh,
			(item.x1+item.x2)/2,
			offset+NODE_TWEEK+NODE_HEIGHT/4+lower,
			html.EscapeString(item.postag))
	}
	fmt.Println("</g>")

	fmt.Printf("<g font-family=\"%s\" font-size=\"%d\" text-anchor=\"middle\" font-style=\"italic\">\n", FONT, int(NODE_FONT_SIZE))
	for _, item := range items {
		enh := ""
		if item.enhanced {
			enh = ` class="enhanced"`
		}
		fmt.Printf("<text%s x=\"%d\" y=\"%d\">%s</text>\n",
			enh,
			(item.x1+item.x2)/2,
			offset-NODE_TWEEK+NODE_HEIGHT*3/4+lower,
			html.EscapeString(item.word))
	}
	fmt.Println("</g>")

	ttips := make([]string, 0)

	fmt.Println("<g opacity=\"0\" stroke-width=\"1\">")
	for i, item := range items {
		enh := ""
		if item.enhanced {
			enh = ` class="enhanced"`
		}
		fmt.Printf("<rect%s x=\"%d\" y=\"%d\" width=\"%d\" height=\"%d\" rx=\"5\" ry=\"5\" "+
			"onclick=\"hold('%s',%d)\" onmouseover=\"mark('%s',%d)\" onmouseout=\"unmark('%s',%d)\" />\n",
			enh,
			item.x1,
			offset,
			item.x2-item.x1,
			int(NODE_HEIGHT),
			svgID,
			i+1,
			svgID,
			i+1,
			svgID,
			i+1)
		ttips = append(ttips, tooltip(item))
	}
	fmt.Println("</g>")

	var boxes bytes.Buffer
	var labels bytes.Buffer
	for _, multi := range multis {
		aa := strings.Split(multi.id, "-")
		if len(aa) != 2 {
			x(fmt.Errorf("Invalid range %s", multi.id), filename, multi.lineno)
		}
		var x1, x2 int
		var found1, found2 bool
		for _, item := range items {
			if aa[0] == item.here {
				x1 = item.x1
				found1 = true
			}
			if aa[1] == item.here {
				x2 = item.x2
				found2 = true
				break
			}
		}
		if !(found1 && found2) {
			x(fmt.Errorf("Invalid range %s", multi.id), filename, multi.lineno)
		}
		fmt.Fprintf(&boxes, "<rect x=\"%d\" y=\"%d\" width=\"%d\" height=\"%d\" rx=\"5\" ry=\"5\" />\n",
			x1,
			offset+NODE_HEIGHT+MULTI_SKIP,
			x2-x1,
			int(MULTI_HEIGHT))
		fmt.Fprintf(&labels, "<text x=\"%d\" y=\"%d\">%s</text>\n",
			(x1+x2)/2,
			offset+NODE_HEIGHT+MULTI_SKIP+MULTI_HEIGHT/2+lower,
			html.EscapeString(multi.word))
	}
	fmt.Printf("<g fill=\"#D0E0FF\" stroke=\"black\" stroke-width=\"1\">\n%s</g>\n", boxes.String())
	fmt.Printf("<g font-family=\"%s\" font-size=\"%d\" font-style=\"italic\" text-anchor=\"middle\">\n%s</g>\n", FONT, int(NODE_FONT_SIZE), labels.String())

	fmt.Printf(`
</svg>
</div>
<script type="text/javascript">
tts['%s'] = [
%s
];`, svgID, strings.Join(ttips, ",\n"))
	if !hasEnhanced {
		fmt.Printf("toggle('%s',false);\n", svgID)
	}
	fmt.Println("</script>")
	if hasEnhanced {
		fmt.Printf(`<div class="udcontrol">
<input type="radio" id="btnb%s" name="btn%s" onclick="toggle('%s',false)" /><label for="btnb%s">Basic</label>
<input type="radio" id="btne%s" name="btn%s" onclick="toggle('%s',true)" checked /><label for="btne%s">Enhanced</label>
</div>
`, svgID, svgID, svgID, svgID, svgID, svgID, svgID, svgID)
	}
}

func anchor(i1, i2, lvl int) float64 {
	a := anchors[i1]
	if len(a) == 1 {
		if i1 < i2 {
			return .75
		}
		return .25
	}
	n := 0
	for i, v := range a {
		if v.dist == i2-i1 && v.level == lvl {
			n = i
			break
		}
	}
	return (float64(n) + .5) / float64(len(a))
}

func classlbl(item *Item) string {
	n := item.end
	uses0 := make(map[int]bool)
	uses1 := make(map[int]bool)
	for _, dep := range dependencies {
		if dep.end == n || dep.headpos == n {
			if dep.rel[0] != "" {
				uses0[dep.end] = true
				uses0[dep.headpos] = true
			}
			if dep.rel[1] != "" {
				uses1[dep.end] = true
				uses1[dep.headpos] = true
			}
		}
	}
	lbls := make([]string, 0, len(uses0)+len(uses1))
	for use := range uses0 {
		lbls = append(lbls, fmt.Sprint("en", use))
	}
	for use := range uses1 {
		lbls = append(lbls, fmt.Sprint("ee", use))
	}
	return strings.Join(lbls, " ")
}

func tooltip(item *Item) string {
	return fmt.Sprintf("['%s','%s','%s','%s','%s']",
		html.EscapeString(item.word),
		html.EscapeString(item.postag),
		html.EscapeString(item.attribs),
		html.EscapeString(item.lemma),
		html.EscapeString(item.xpostag))
}

func textwidth(text string, fontsize float64, bold bool) (width, height, lift int) {

	var sizes []uint8
	var asc, desc int
	if bold {
		sizes = fontBoldSizes
		asc = fontBoldAscent
		desc = fontBoldDescent
	} else {
		sizes = fontRegularSizes
		asc = fontRegularAscent
		desc = fontRegularDescent
	}

	w := 0
	for _, c := range text {
		i := int(c)
		var w1 int
		if i >= len(sizes) {
			w1 = fontBaseSize
		} else {
			w1 = int(sizes[i])
		}
		w += w1
	}
	return int(fontsize * float64(w) / float64(fontBaseSize)),
		int(fontsize * float64(asc+desc) / float64(fontBaseSize)),
		int(fontsize * float64(asc) / float64(fontBaseSize))
}

func max(a int, b ...int) int {
	for _, i := range b {
		if i > a {
			a = i
		}
	}
	return a
}

func abs(i int) int {
	if i < 0 {
		return -i
	}
	return i
}

func x(err error, filename string, lineno int) {
	if err == nil {
		return
	}
	if lineno > 0 {
		fmt.Fprintf(os.Stderr, "in %s, line %d: %v\n", filename, lineno, err)
	} else {
		fmt.Fprintln(os.Stderr, err)
	}
	os.Exit(1)
}
