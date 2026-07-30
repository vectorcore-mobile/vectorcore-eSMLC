#!/usr/bin/env python3
"""Formatting-only normalization of v2 WordprocessingML-derived ASN.1."""
import hashlib,json
from pathlib import Path
R=Path(__file__).resolve().parents[4]; I=R/'docs/specs/asn1/lpp/r16.4.0/modules_v2/original'; O=R/'docs/specs/asn1/lpp/r16.4.0/modules_v2/normalized'
def sha(p): return hashlib.sha256(p.read_bytes()).hexdigest()
def main():
 O.mkdir(parents=True,exist_ok=True); records=[]
 for p in sorted(I.glob('*.asn1')):
  old=p.read_text(); new='\n'.join(line.replace('\t','    ').rstrip() for line in old.splitlines())+'\n'
  q=O/p.name;q.write_text(new)
  records.append({'module':p.stem,'original_sha256':sha(p),'normalized_sha256':sha(q),'changes':[{'rule':'tabs_to_four_spaces_and_trailing_whitespace','semantic_impact':'none'}]})
 (O.parent/'normalization-manifest.json').write_text(json.dumps({'records':records},indent=2)+'\n')
 (O.parent/'normalization-report.md').write_text('# V2 normalization\n\nOnly XML-derived TAB rendering was changed to four spaces and trailing whitespace removed. No ASN.1 token, field, constraint, tag, or ordering changed.\n')
if __name__=='__main__':main()
