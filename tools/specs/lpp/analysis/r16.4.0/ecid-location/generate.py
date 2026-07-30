#!/usr/bin/env python3
"""Generate deterministic ECID location analysis fixtures with pinned UPER."""
import hashlib,json
from pathlib import Path
import asn1tools
from asn1tools.codecs.uper import Encoder
R=Path(__file__).resolve().parents[6]; P=R/'docs/specs/asn1/lpp/r16.4.0/modules_v2/normalized/LPP-PDU-Definitions.asn1'; O=R/'tools/specs/lpp/analysis/r16.4.0/ecid-location'; F=R/'tools/specs/lpp/fixtures/r16.4.0/ecid-location'
def dump(p,x):p.parent.mkdir(parents=True,exist_ok=True);p.write_text(json.dumps(x,indent=2,sort_keys=True)+'\n')
def enc(s,t,v):
 b=bytes(s.encode(t,v));e=Encoder();s.types[t]._type.encode(v,e);return b,e.chunks_number_of_bits+e.number_of_bits
def main():
 s=asn1tools.compile_files([str(P)],codec='uper')
 req=lambda bits:{'criticalExtensions':('c1',('requestLocationInformation-r9',{'ecid-RequestLocationInformation':{'requestedMeasurements':bits}}))}
 elem={'physCellId':1,'arfcnEUTRA':100,'rsrp-Result':30}
 prov={'criticalExtensions':('c1',('provideLocationInformation-r9',{'ecid-ProvideLocationInformation':{'ecid-SignalMeasurementInformation':{'measuredResultsList':[elem]}}}))}
 vals={'request-ecid-rsrp':('RequestLocationInformation',req((bytes([128]),1))),'request-ecid-all-root':('RequestLocationInformation',req((bytes([224]),3))),'provide-ecid-rsrp-one-cell':('ProvideLocationInformation',prov)};rows=[]
 for n,(t,v) in vals.items():
  b,k=enc(s,t,v);d=s.decode(t,b);b2,k2=enc(s,t,d);assert b==b2 and k==k2;(F/'valid').mkdir(parents=True,exist_ok=True);(F/'valid'/(n+'.uper')).write_bytes(b);rows.append({'name':n,'type':t,'hex':b.hex(),'bit_length':k,'unused_trailing_bits':len(b)*8-k,'sha256':hashlib.sha256(b).hexdigest(),'round_trip':True})
 dump(F/'manifest.json',{'compiler':'asn1tools','compiler_version':asn1tools.__version__,'encoding_rule':'uper','fixtures':rows,'negative_plan':['requestedMeasurements length 0/9','missing measuredResultsList','physCellId >503','arfcnEUTRA out of range','truncation','extension present']})
 dump(O/'field-inventory.json',{'request_r9_root':['common','a-gnss','otdoa','ecid','epdu'],'provide_r9_root':['common','a-gnss','otdoa','ecid','epdu'],'ecid_request':[{'name':'requestedMeasurements','type':'BIT STRING SIZE(1..8)','mandatory':True}], 'ecid_provide':[{'name':'ecid-SignalMeasurementInformation','optional':True},{'name':'ecid-Error','optional':True}], 'measurement_element':[{'name':'physCellId','type':'INTEGER(0..503)','mandatory':True},{'name':'cellGlobalId','type':'CellGlobalIdEUTRA-AndUTRA','optional':True},{'name':'arfcnEUTRA','type':'INTEGER','mandatory':True},{'name':'systemFrameNumber','type':'BIT STRING SIZE(10)','optional':True},{'name':'rsrp-Result','type':'INTEGER(0..97)','optional':True},{'name':'rsrq-Result','type':'INTEGER(0..34)','optional':True},{'name':'ue-RxTxTimeDiff','type':'INTEGER(0..4095)','optional':True}]})
 dump(O/'primitive-gaps.json',{'available':['BOOLEAN','constrained INTEGER','root CHOICE','optional bitmap','bounded variable BIT STRING'], 'missing':['SIZE-constrained SEQUENCE OF (1..32)','fixed BIT STRING SIZE(10)','CellGlobalIdEUTRA-AndUTRA nested PLMN digit lists','ECID error CHOICE/ENUMERATED extension handling'], 'initial_request_only_requires':['bounded BIT STRING SIZE(1..8)','sequence extension bit','R9 five-bit optional bitmap']})
 dump(O/'resource-bounds.json',{'measuredResultsList_max':32,'request_bitstring_max':8,'measurement_depth':4,'initial_request_no_lists':True,'provide_initially_defer':'all measurement lists, CellGlobalId, systemFrameNumber and error CHOICE'})
if __name__=='__main__':main()
