#!/usr/bin/env python3
"""Deterministically assemble contiguous ASN.1 modules from extracted blocks."""
import hashlib,json,re,os
from pathlib import Path
R=Path(__file__).resolve().parents[4]; v=os.environ.get('LPP_SOURCE_VERSION','v1'); B=R/('docs/specs/asn1/lpp/r16.4.0/source_v2/original' if v=='v2' else 'docs/specs/asn1/lpp/r16.4.0/source/original'); O=R/('docs/specs/asn1/lpp/r16.4.0/modules_v2/original' if v=='v2' else 'docs/specs/asn1/lpp/r16.4.0/modules/original')
def sha(p): return hashlib.sha256(p.read_bytes()).hexdigest()
def main():
 files=sorted(B.glob('block-*.asn1')); starts=[]
 for i,p in enumerate(files):
  t=p.read_text()
  if 'DEFINITIONS AUTOMATIC TAGS ::=' in t:
   m=re.search(r'([A-Za-z][A-Za-z0-9-]*)\s*(?:\{|\n)\s*(?:[^\n]*\n){0,4}?DEFINITIONS AUTOMATIC TAGS ::=',t)
   if not m: raise SystemExit('unparseable header '+p.name)
   starts.append((i,m.group(1)))
 if len(starts)!=2: raise SystemExit('expected 2 modules, found %d'%len(starts))
 O.mkdir(parents=True,exist_ok=True); out=[]
 for n,(a,name) in enumerate(starts):
  z=starts[n+1][0] if n+1<len(starts) else len(files); chunk=files[a:z]; text='\n\n'.join(p.read_text().rstrip() for p in chunk)+'\n'
  if not text.rstrip().endswith('END'): raise SystemExit('missing END '+name)
  q=O/(name+'.asn1');q.write_text(text);out.append({'module':name,'blocks':[p.name for p in chunk],'sha256':sha(q)})
 (O.parent/'original-manifest.json').write_text(json.dumps({'modules':out},indent=2)+'\n')
 print('assembled',len(out),'modules')
if __name__=='__main__':main()
