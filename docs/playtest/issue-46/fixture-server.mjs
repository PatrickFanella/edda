import http from 'node:http';
import { WebSocketServer } from 'ws';
const timestamp = '2026-01-01T12:00:00Z';
const user = {id:'fixture-user', name:'Fixture Player', email:'player@example.test'};
const campaign = {id:'alpha',name:'Adventure alpha',description:'An isolated adventure',genre:'fantasy',tone:'hopeful',themes:['discovery'],status:'active',rules_mode:'narrative',created_by:user.id,created_at:timestamp,updated_at:timestamp};
const character = {id:'hero',campaign_id:'alpha',user_id:user.id,name:'Mira',description:'A scout',stats:{},hp:10,max_hp:10,experience:0,level:1,status:'active',abilities:[]};
const history = [{turn_number:1,player_input:'',input_type:'narrative',llm_response:'The alpha gate opens.',choices:['Enter the gate'],created_at:timestamp},{turn_number:2,player_input:'Enter the gate',input_type:'action',llm_response:'The alpha keeper welcomes you.',created_at:timestamp}];
const saves = [{id:'save-1',campaign_id:'alpha',name:'Fixture checkpoint',turn_number:2,is_auto:false,created_at:timestamp}];
const arrays = ['/quests','/character/inventory','/character/abilities','/character/feats','/character/skills','/npcs','/npcs/encountered','/journal/summaries','/journal/entries','/locations','/facts','/journal','/encountered','/relationships','/codex/languages','/codex/cultures','/codex/beliefs','/codex/economies'];
const server = http.createServer(async (req,res) => {
 const path = new URL(req.url,'http://fixture').pathname.replace('/api/v1','');
 console.log(req.method, path);
 let body, status=200;
 if(path==='/auth/me') body={user};
 else if(path==='/auth/login') body={token:'fixture-token',user};
 else if(path==='/auth/logout') body={};
 else if(path==='/campaigns') body={campaigns:[campaign]};
 else if(path==='/campaigns/alpha') body=campaign;
 else if(path==='/campaigns/alpha/history') body={entries:history};
 else if(path==='/campaigns/alpha/character') body=character;
 else if(path==='/campaigns/alpha/map') body={locations:[],connections:[]};
 else if(path==='/campaigns/alpha/time') body={day:1,hour:12,minute:0};
 else if(path==='/campaigns/alpha/saves') {
   if(req.method==='POST') {
     let raw=''; for await (const part of req) raw+=part;
     saves.push({...saves[0],id:'save-'+(saves.length+1),name:JSON.parse(raw).name});
     body=saves.at(-1);
   } else body=saves;
 } else if(path.startsWith('/campaigns/alpha/export/')) {
   const format=path.split('/').at(-1);
   res.writeHead(200,{'Content-Type':format==='json'?'application/json':'text/markdown'});
   res.end(format==='json'?JSON.stringify({campaign,entries:history}):'# Fixture export\nThe alpha gate opens.\n'); return;
 } else if(arrays.includes(path.replace('/campaigns/alpha',''))) body=[];
 else {status=501;body={error:'Unmapped isolated fixture: '+path};}
 res.writeHead(status,{'Content-Type':'application/json'});res.end(JSON.stringify(body));
});
const ws = new WebSocketServer({server,path:'/api/v1/campaigns/alpha/ws'});
ws.on('connection', socket => socket.on('message', raw => {
 const input=JSON.parse(String(raw)).payload.input; console.log('ACTION',input);
 if(input==='fixture error') {socket.send(JSON.stringify({type:'error',timestamp,payload:{error:'Fixture validation error: try a different action.'}})); return;}
 const long=input==='long choices';
 socket.send(JSON.stringify({type:'result',timestamp,payload:{narrative:long?'A long fixture scene. '.repeat(180):'You find a silver key.',state_changes:[],resolution_events:[],choices:long?Array.from({length:8},(_,i)=>({id:'long-'+i,text:'Choice '+(i+1)+': '+ 'Inspect the ancient inscription before entering the crowded chamber. '.repeat(4)})):[{id:'inspect',text:'Inspect the key'}]}}));
}));
server.listen(39246,'127.0.0.1',()=>console.log('Fixture-only HTTP + WebSocket API on 39246; PID',process.pid));
