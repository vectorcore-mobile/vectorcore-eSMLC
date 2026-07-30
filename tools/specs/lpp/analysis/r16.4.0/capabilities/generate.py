#!/usr/bin/env python3
"""Deterministically inventory and fixture the bounded R9 ECID capability plan."""
from __future__ import annotations
import hashlib, json
from pathlib import Path
import asn1tools
from asn1tools.codecs.uper import Encoder

ROOT = Path(__file__).resolve().parents[6]
PDU = ROOT / "docs/specs/asn1/lpp/r16.4.0/modules_v2/normalized/LPP-PDU-Definitions.asn1"
GRAPH = ROOT / "docs/specs/asn1/lpp/r16.4.0/modules_v2/analysis/reference-graph.json"
OUT = ROOT / "tools/specs/lpp/analysis/r16.4.0/capabilities"
FIX = ROOT / "tools/specs/lpp/fixtures/r16.4.0/capabilities"

def sha(p): return hashlib.sha256(p.read_bytes()).hexdigest()
def dump(p, value): p.parent.mkdir(parents=True, exist_ok=True); p.write_text(json.dumps(value, indent=2, sort_keys=True)+"\n")
def jsonable(v):
    if isinstance(v, tuple): return [jsonable(x) for x in v]
    if isinstance(v, bytes): return {"hex":v.hex()}
    if isinstance(v, dict): return {k:jsonable(x) for k,x in v.items()}
    if isinstance(v, list): return [jsonable(x) for x in v]
    return v
def bits(spec, typ, value):
    encoded=bytes(spec.encode(typ,value,check_constraints=True)); enc=Encoder(); spec.types[typ]._type.encode(value,enc)
    n=enc.chunks_number_of_bits+enc.number_of_bits
    if bytes(enc.as_bytearray())!=encoded: raise RuntimeError("public/instrumented encoder mismatch")
    return encoded,n
def closure(root, graph):
    seen=set(); todo=[root]
    while todo:
        x=todo.pop()
        if x not in seen: seen.add(x);todo.extend(graph.get(x,[]))
    return sorted(seen)
def main():
    g={n:[] for n in json.loads(GRAPH.read_text())["nodes"]}
    for e in json.loads(GRAPH.read_text())["edges"]: g[e["from"]].append(e["to"])
    request=closure("RequestCapabilities-r9-IEs",g); provide=closure("ProvideCapabilities-r9-IEs",g)
    selected=sorted(set(closure("ECID-RequestCapabilities",g)+closure("ECID-ProvideCapabilities",g)+["RequestCapabilities-r9-IEs","ProvideCapabilities-r9-IEs","RequestCapabilities","ProvideCapabilities"]))
    definitions={"source_module_sha256":sha(PDU),"request_path":["RequestCapabilities","criticalExtensions:c1[0]","requestCapabilities-r9[0]","RequestCapabilities-r9-IEs"],"provide_path":["ProvideCapabilities","criticalExtensions:c1[0]","provideCapabilities-r9[0]","ProvideCapabilities-r9-IEs"],"request_fields":[
      ["commonIEsRequestCapabilities","CommonIEsRequestCapabilities","optional root","deferred"],["a-gnss-RequestCapabilities","A-GNSS-RequestCapabilities","optional root","deferred"],["otdoa-RequestCapabilities","OTDOA-RequestCapabilities","optional root","deferred"],["ecid-RequestCapabilities","ECID-RequestCapabilities","optional root","selected"],["epdu-RequestCapabilities","EPDU-Sequence","optional root","deferred"],
      ["sensor/tbs/wlan/bt r13","various","extension additions","rejected"],["nr-* r16","various","extension additions","rejected"]],"provide_fields":[
      ["commonIEsProvideCapabilities","CommonIEsProvideCapabilities","optional root","deferred"],["a-gnss-ProvideCapabilities","A-GNSS-ProvideCapabilities","optional root","deferred"],["otdoa-ProvideCapabilities","OTDOA-ProvideCapabilities","optional root","deferred"],["ecid-ProvideCapabilities","ECID-ProvideCapabilities","optional root","selected"],["epdu-ProvideCapabilities","EPDU-Sequence","optional root","deferred"],
      ["sensor/tbs/wlan/bt r13","various","extension additions","rejected"],["nr-* r16","various","extension additions","rejected"]],"selected_ecid":{"request":"ECID-RequestCapabilities ::= SEQUENCE { ... }","provide_field":"ecid-MeasSupported BIT STRING {rsrpSup(0), rsrqSup(1), ueRxTxSup(2), nrsrpSup-r14(3), nrsrqSup-r14(4)} (SIZE(1..8))","ownership":"ProvideCapabilities reports target-device measurement support; this phase makes no location-server capability claim."}}
    dump(OUT/"definitions.json",definitions)
    closures={"complete_r9":{"request_definitions":request,"provide_definitions":provide,"definition_count":len(set(request+provide)),"unresolved":[]},"recommended_ecid":{"definitions":selected,"definition_count":len(selected),"unresolved":[],"reason":"ECID advertises a useful broad measurement family with one bounded mandatory bit string; OTDOA adds bounded band lists and A-GNSS adds 50 reachable support definitions."},"minimum_fixture":{"definitions":["RequestCapabilities","RequestCapabilities-r9-IEs","ECID-RequestCapabilities","ProvideCapabilities","ProvideCapabilities-r9-IEs","ECID-ProvideCapabilities"],"definition_count":6,"unresolved":[]}}
    for v in closures.values(): v["sha256"]=hashlib.sha256(json.dumps(v,sort_keys=True,separators=(",",":" )).encode()).hexdigest()
    dump(OUT/"closures.json",closures); dump(OUT/"dependencies.json",{"graph_source_sha256":sha(GRAPH),"closures":closures})
    dump(OUT/"uper-requirements.json",{"sufficient_existing":["optional bitmap","root CHOICE","extension-presence bit","BOOLEAN"],"missing":[{"primitive":"bounded variable-size BIT STRING","field":"ECID-ProvideCapabilities.ecid-MeasSupported","constraint":"SIZE(1..8)","uper_requirement":"constrained length followed by exactly that many MSB-first bits","proposed_api":"ReadBitString(min,max) / WriteBitString(bits,min,max)","required":True}],"deferred":["SEQUENCE OF and band constraints (OTDOA)","nested lists/bit strings (A-GNSS)","extension additions and open payloads"]})
    plan={"positive":["request-empty","request-ecid-selector","provide-empty","provide-ecid-rsrp","provide-ecid-rsrp-rsrq-uerxtx"],"negative":["ecid-bitstring-empty","ecid-bitstring-too-long","invalid-critical-extension-choice","non-zero-padding-after-encoding"],"selected_fields":["ecid-RequestCapabilities","ecid-ProvideCapabilities.ecid-MeasSupported"]};dump(OUT/"fixture-plan.json",plan);dump(OUT/"field-inventory.json",definitions)
    spec=asn1tools.compile_files([str(PDU)],codec="uper")
    values={
      "request-empty":("RequestCapabilities",{"criticalExtensions":("c1",("requestCapabilities-r9",{}))},"R9 empty request remains valid"),
      "request-ecid-selector":("RequestCapabilities",{"criticalExtensions":("c1",("requestCapabilities-r9",{"ecid-RequestCapabilities":{}}))},"requests ECID capability group"),
      "provide-empty":("ProvideCapabilities",{"criticalExtensions":("c1",("provideCapabilities-r9",{}))},"R9 empty provide remains valid"),
      "provide-ecid-rsrp":("ProvideCapabilities",{"criticalExtensions":("c1",("provideCapabilities-r9",{"ecid-ProvideCapabilities":{"ecid-MeasSupported":(bytes([128]),1)}}))},"target supports ECID RSRP"),
      "provide-ecid-rsrp-rsrq-uerxtx":("ProvideCapabilities",{"criticalExtensions":("c1",("provideCapabilities-r9",{"ecid-ProvideCapabilities":{"ecid-MeasSupported":(bytes([224]),3)}}))},"target supports three root ECID measurements")}
    rec=[]
    for name,(typ,value,purpose) in values.items():
      data,n=bits(spec,typ,value);decoded=spec.decode(typ,data);again,n2=bits(spec,typ,decoded)
      if data!=again or n!=n2 or jsonable(decoded)!=jsonable(value):raise RuntimeError(name+" round trip")
      (FIX/"valid").mkdir(parents=True,exist_ok=True);(FIX/"canonical").mkdir(parents=True,exist_ok=True)
      (FIX/"valid"/(name+".uper")).write_bytes(data);dump(FIX/"canonical"/(name+".json"),jsonable(value))
      rec.append({"name":name,"top_level_type":typ,"purpose":purpose,"binary_file":"valid/"+name+".uper","canonical_file":"canonical/"+name+".json","hex":data.hex(),"bit_length":n,"unused_trailing_bits":len(data)*8-n,"sha256":hashlib.sha256(data).hexdigest(),"decoded":jsonable(decoded),"round_trip":"byte-and-bit-identical"})
    negatives=[]
    for name,v in {"ecid-bitstring-empty":(bytes(),0),"ecid-bitstring-too-long":(bytes([255,128]),9)}.items():
      value={"criticalExtensions":("c1",("provideCapabilities-r9",{"ecid-ProvideCapabilities":{"ecid-MeasSupported":v}}))}
      try: bits(spec,"ProvideCapabilities",value); ok=False
      except Exception as e: ok=True; negatives.append({"name":name,"rejected":True,"error_type":type(e).__name__})
      if not ok: raise RuntimeError(name+" accepted")
    manifest={"specification":"3GPP TS 37.355 V16.4.0","encoding_rule":"uper","compiler":"asn1tools","compiler_version":asn1tools.__version__,"module_sha256":sha(PDU),"closure_sha256":closures["recommended_ecid"]["sha256"],"fixtures":rec,"negative_cases":negatives,"bit_length_method":"asn1tools Encoder pre-octet-padding"};dump(FIX/"manifest.json",manifest)
    (FIX/"README.md").write_text("# Independent bounded ECID capability UPER fixtures\n\nGenerated only by the pinned development-only asn1tools workflow; no runtime package reads these files.\n")
    print("generated",len(rec),"valid and",len(negatives),"negative fixtures")
if __name__=="__main__":main()
