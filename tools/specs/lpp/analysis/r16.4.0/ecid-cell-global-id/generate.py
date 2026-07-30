#!/usr/bin/env python3
import hashlib,json
from pathlib import Path
import asn1tools
from asn1tools.codecs.per import Encoder
R=Path(__file__).resolve().parent; H=R/'cell-global-id-harness.asn1'; O=R/'fixtures.json'; S=asn1tools.compile_files([str(H)],codec='uper')
def bitlen(v):
 e=Encoder();S.types['CellGlobalIdEUTRA-AndUTRA']._type.encode(v,e);return e.chunks_number_of_bits+e.number_of_bits
def cell(mcc,mnc,k,n):
 bits=28 if k=='eutra' else 32
 raw=n << ((8-bits%8)%8)
 return {'plmn-Identity':{'mcc':mcc,'mnc':mnc},'cellIdentity':(k,(raw.to_bytes((bits+7)//8,'big'),bits))}
def rec(name,v):
 b=S.encode('CellGlobalIdEUTRA-AndUTRA',v);d=S.decode('CellGlobalIdEUTRA-AndUTRA',b);assert S.encode('CellGlobalIdEUTRA-AndUTRA',d)==b
 return {'name':name,'hex':b.hex(),'bit_length':bitlen(v),'sha256':hashlib.sha256(b).hexdigest(),'round_trip':True}
vals=[('eutra-001-01-zero',cell([0,0,1],[0,1],'eutra',0)),('eutra-001-01-one',cell([0,0,1],[0,1],'eutra',1)),('eutra-310-260-pattern',cell([3,1,0],[2,6,0],'eutra',0x1234567)),('eutra-max',cell([3,1,0],[2,6,0],'eutra',(1<<28)-1)),('utra-001-01-zero',cell([0,0,1],[0,1],'utra',0)),('utra-001-01-one',cell([0,0,1],[0,1],'utra',1)),('utra-310-260-pattern',cell([3,1,0],[2,6,0],'utra',0x12345678)),('utra-max',cell([3,1,0],[2,6,0],'utra',(1<<32)-1))]
O.write_text(json.dumps({'compiler':'asn1tools','compiler_version':asn1tools.__version__,'encoding_rule':'uper','harness_sha256':hashlib.sha256(H.read_bytes()).hexdigest(),'fixtures':[rec(n,v) for n,v in vals]},indent=2,sort_keys=True)+'\n')
