import pdfplumber
import json
import re

pdf_path = "/home/sko/projects/go-pdf-test/testdata/bg_01.pdf"

with pdfplumber.open(pdf_path) as pdf:
    page = pdf.pages[0]
    page_height = page.height
    page_width = page.width

    print(f"Page size: {page_width} x {page_height}")
    print()

    words = page.extract_words(extra_attrs=["fontname", "size"])

    print(f"=== ALL WORDS ({len(words)}) ===")
    for i, w in enumerate(words):
        has_dots = bool(re.search(r'[.…]{2,}', w['text']))
        marker = ">" if has_dots else " "
        print(f"{marker} word[{i:2d}]: {w['text'][:50]:50s}  x0={w['x0']:6.1f} top={w['top']:6.1f} x1={w['x1']:6.1f} w={w['x1']-w['x0']:6.1f}  font={w['fontname']}")

    print()
    print("=== DOT PLACEHOLDERS ===")
    idx = 0
    for w in words:
        text = w['text']
        if re.search(r'[.…]{4,}', text):
            print(f"  id={idx:2d}  x0={w['x0']:6.1f}  top={w['top']:6.1f}  x1={w['x1']:6.1f}  w={w['x1']-w['x0']:6.1f}  text={text[:40]}")
            idx += 1
