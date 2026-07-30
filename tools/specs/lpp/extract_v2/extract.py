#!/usr/bin/env python3
"""Version-2 WordprocessingML extractor: preserves text, tabs, and breaks."""
import importlib.util
from pathlib import Path

# Reuse v1's integrity/marker pipeline, replacing only paragraph semantics.
root=Path(__file__).resolve().parents[4]
spec=importlib.util.spec_from_file_location('v1',root/'tools/specs/lpp/extract/extract.py');v1=importlib.util.module_from_spec(spec);spec.loader.exec_module(v1)
def paragraph(p):
 out=[]
 for e in p.iter():
  tag=e.tag.rsplit('}',1)[-1]
  if tag=='t': out.append(e.text or '')
  elif tag=='tab': out.append('\t')
  elif tag in ('br','cr'): out.append('\n')
 return ''.join(out)
v1.paragraph=paragraph
v1.OUT=root/'docs/specs/asn1/lpp/r16.4.0/source_v2/original'
v1.MANIFEST=root/'docs/specs/asn1/lpp/r16.4.0/source_v2/manifest.json'
v1.main()
