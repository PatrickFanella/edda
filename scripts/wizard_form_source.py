"""Complete typed source dispatch inventory; execution bindings are separate."""
import json,re,subprocess,hashlib
from pathlib import Path
PIN='8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87'
ROOT=Path(__file__).resolve().parents[1]
def inventory(pdf):
 assert hashlib.sha256(pdf.read_bytes()).hexdigest()==PIN
 source=json.loads((ROOT/'docs/investigations/issue-82-world-inventory.json').read_text())['records']
 source=[r for r in source if r['family']=='monsters'];names={r['name'] for r in source}
 lines=[]
 for page in range(258,365):
  for x in [0,297]:
   text=subprocess.check_output(['pdftotext','-f',str(page),'-l',str(page),'-x',str(x),'-y','0','-W','297','-H','783','-layout',str(pdf),'-'],text=True)
   lines += [(page,' '.join(t.split())) for t in text.splitlines() if t.strip() and 'System Reference Document' not in t]
 starts={}
 for i,(page,text) in enumerate(lines[:-1]):
  if text in names and re.match(r'(Tiny|Small|Medium|Large|Huge|Gargantuan)',lines[i+1][1]):starts[text]=i
 answer=[]
 for r in source:
  assert r['name'] in starts,('missing monster header',r['name'])
  i=starts[r['name']];following=next((j for j in range(i+1,len(lines)) if lines[j][1] in starts and starts[lines[j][1]]==j),len(lines))
  block='\n'.join(t for _,t in lines[i:following]);cr=re.search(r'\bCR (\d+(?:/\d+)?)\b',block)
  assert cr,('missing CR',r['name'])
  kind=re.search(r'\b(Aberration|Beast|Celestial|Construct|Dragon|Elemental|Fey|Fiend|Giant|Humanoid|Monstrosity|Ooze|Plant|Undead)\b',lines[i+1][1]);assert kind,r['name']
  answer.append(dict(id=r['id'],name=r['name'],page=lines[i][0],type=kind[1],cr=cr[1],typed_output_id='srd-5.2.1:execution/monster/'+r['id'].rsplit('/',1)[-1],shapechange_source_eligible_type=kind[1] not in ['Construct','Undead'],simulacrum_source_eligible_type=kind[1] in ['Beast','Humanoid'],execution_status='UNBOUND until exact qualified output binding',header=lines[i+1][1]))
 return answer
if __name__=='__main__':
 import sys
 rows=inventory(Path(sys.argv[1]));Path(sys.argv[2]).write_text(json.dumps(rows,indent=2)+'\n');print(len(rows),'typed form identities; no execution implied')
