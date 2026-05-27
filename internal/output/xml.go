package output

import (
	"encoding/xml"
	"strings"
)

// XMLRenderer produces a structured XML document.
type XMLRenderer struct{}

// FileExtension returns ".xml".
func (XMLRenderer) FileExtension() string { return ".xml" }

// Render returns the document as an XML string. The output is hand-rolled
// rather than reflection-driven because we want CDATA-wrapped file content
// and stable section ordering.
func (XMLRenderer) Render(doc *Document) (string, error) {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	b.WriteString(`<gophercram>` + "\n")

	if doc.Config.Output.FileSummary {
		writeXMLSummary(&b, doc)
	}
	if doc.HeaderText != "" {
		b.WriteString("  <user_header>")
		writeEscaped(&b, doc.HeaderText)
		b.WriteString("</user_header>\n")
	}
	if doc.Config.Output.DirectoryStructure {
		b.WriteString("  <directory_structure>")
		writeCDATA(&b, doc.DirectoryTree)
		b.WriteString("</directory_structure>\n")
	}
	if doc.Config.Output.Files && len(doc.Files) > 0 {
		b.WriteString("  <files>\n")
		for _, f := range doc.Files {
			if f.Skipped {
				continue
			}
			b.WriteString("    <file path=\"")
			b.WriteString(xmlAttr(f.RelPath))
			b.WriteString("\">")
			writeCDATA(&b, f.Content)
			b.WriteString("</file>\n")
		}
		b.WriteString("  </files>\n")
	}
	if doc.GitDiffWorkTree != "" || doc.GitDiffStaged != "" {
		b.WriteString("  <git_diffs>\n")
		if doc.GitDiffWorkTree != "" {
			b.WriteString("    <work_tree>")
			writeCDATA(&b, doc.GitDiffWorkTree)
			b.WriteString("</work_tree>\n")
		}
		if doc.GitDiffStaged != "" {
			b.WriteString("    <staged>")
			writeCDATA(&b, doc.GitDiffStaged)
			b.WriteString("</staged>\n")
		}
		b.WriteString("  </git_diffs>\n")
	}
	if len(doc.GitLog) > 0 {
		b.WriteString("  <git_log>\n")
		for _, c := range doc.GitLog {
			b.WriteString("    <commit hash=\"")
			b.WriteString(xmlAttr(c.Hash))
			b.WriteString("\">\n")
			b.WriteString("      <author>")
			writeEscaped(&b, c.Author)
			b.WriteString("</author>\n")
			b.WriteString("      <date>")
			writeEscaped(&b, c.Date)
			b.WriteString("</date>\n")
			b.WriteString("      <message>")
			writeEscaped(&b, c.Message)
			b.WriteString("</message>\n")
			if len(c.Files) > 0 {
				b.WriteString("      <files>\n")
				for _, f := range c.Files {
					b.WriteString("        <file>")
					writeEscaped(&b, f)
					b.WriteString("</file>\n")
				}
				b.WriteString("      </files>\n")
			}
			b.WriteString("    </commit>\n")
		}
		b.WriteString("  </git_log>\n")
	}
	if doc.Instruction != "" {
		b.WriteString("  <instruction>")
		writeCDATA(&b, doc.Instruction)
		b.WriteString("</instruction>\n")
	}
	b.WriteString("</gophercram>\n")
	return b.String(), nil
}

func writeXMLSummary(b *strings.Builder, doc *Document) {
	b.WriteString("  <summary>\n")
	b.WriteString("    <producer>")
	writeEscaped(b, doc.AppName+" v"+doc.AppVersion)
	b.WriteString("</producer>\n")
	b.WriteString("    <generated_at>")
	writeEscaped(b, doc.GeneratedAt.UTC().Format("2006-01-02T15:04:05Z"))
	b.WriteString("</generated_at>\n")
	b.WriteString("    <description>This file packs the contents of a software repository into a single document so it can be supplied as context to a language model.</description>\n")
	b.WriteString("    <stats files=\"")
	b.WriteString(itoa(doc.Aggregate.TotalFiles))
	b.WriteString("\" chars=\"")
	b.WriteString(itoa(doc.Aggregate.TotalChars))
	b.WriteString("\" tokens=\"")
	b.WriteString(itoa(doc.Aggregate.TotalTokens))
	b.WriteString("\" bytes=\"")
	b.WriteString(itoa64(doc.Aggregate.TotalBytes))
	b.WriteString("\"/>\n")
	b.WriteString("  </summary>\n")
}

// writeCDATA writes content wrapped in a CDATA section, splitting if the
// content contains the `]]>` terminator.
func writeCDATA(b *strings.Builder, content string) {
	if content == "" {
		return
	}
	if !strings.Contains(content, "]]>") {
		b.WriteString("<![CDATA[")
		b.WriteString(content)
		b.WriteString("]]>")
		return
	}
	parts := strings.Split(content, "]]>")
	for i, p := range parts {
		b.WriteString("<![CDATA[")
		b.WriteString(p)
		if i < len(parts)-1 {
			b.WriteString("]]]]><![CDATA[>")
		}
		b.WriteString("]]>")
	}
}

func writeEscaped(b *strings.Builder, s string) {
	_ = xml.EscapeText(strWriter{b}, []byte(s))
}

// xmlAttr escapes a string for use as an XML attribute value (always quoted
// with `"`). Quotes inside the value are replaced with the entity.
func xmlAttr(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

// strWriter adapts a strings.Builder to io.Writer.
type strWriter struct{ b *strings.Builder }

func (w strWriter) Write(p []byte) (int, error) { return w.b.Write(p) }
