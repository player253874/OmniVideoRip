const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const vm = require('node:vm');
const source = fs.readFileSync(require('node:path').join(__dirname, 'src/task-snapshots.js'), 'utf8');
function makeSync() { const ctx = vm.createContext({}); vm.runInContext(source,ctx);return new ctx.TaskSnapshots(new Map()); }

test('clear rejects delayed response and late progress; new requests still work', () => {
 const sync=makeSync(); sync.tasks.set('old',{id:'old'});
 const old=sync.request(); sync.clear();
 assert.equal(sync.apply(old,[{id:'old'}]),false);
 assert.equal(sync.progress({id:'old',status:'completed'}),false);
 assert.equal(sync.tasks.size,0);
 assert.equal(sync.apply(sync.request(),[{id:'new'}]),true);
 assert.equal(sync.tasks.size,1);
});
test('empty authoritative snapshot removes ghosts; out of order snapshots ignored',()=>{
 const sync=makeSync(); const first=sync.request(),second=sync.request();
 sync.apply(second,[{id:'new'}]); assert.equal(sync.apply(first,[{id:'old'}]),false);
 sync.apply(sync.request(),[]); assert.equal(sync.tasks.size,0);
});
test('deletion invalidates snapshots and late completion events',()=>{
 const sync=makeSync();sync.tasks.set('one',{id:'one'});const pending=sync.request();
 sync.remove('one');assert.equal(sync.apply(pending,[{id:'one'}]),false);
 assert.equal(sync.progress({id:'one'}),false);
});

function appFixture(api) {
 const handlers = {}, alerts=[];
 const button={disabled:false,addEventListener:(name,fn)=>{handlers[name]=fn}};
 const ctx=vm.createContext({
  window:{go:{main:{App:api}},alert:value=>alerts.push(value)},
  document:{getElementById:id=>id==='clear-queue-btn'?button:null,addEventListener:()=>{},querySelectorAll:()=>[]},
  localStorage:{getItem:()=>null,setItem:()=>{},removeItem:()=>{}},
  setInterval:()=>1,clearInterval:()=>{},console:{warn:()=>{}},
 });
 vm.runInContext(source,ctx);
 vm.runInContext(fs.readFileSync(require('node:path').join(__dirname,'src/app.js'),'utf8'),ctx);
 vm.runInContext('bindEvents()',ctx);
 return {ctx,button,handlers,alerts,tasks:vm.runInContext('state.tasksMap',ctx)};
}

test('actual clear button rejects a pre-clear GetAllTasks response',async()=>{
 let resolve;
 const app=appFixture({GetAllTasks:()=>new Promise(r=>resolve=r),ClearAllTasksAndCaches:async()=>{}});
 app.tasks.set('old',{id:'old',status:'paused'});
 const pending=vm.runInContext('refreshTaskSnapshot()',app.ctx);
 await app.handlers.click();
 resolve([{id:'old',status:'paused'}]);await pending;
 assert.equal(app.tasks.size,0);
 assert.equal(app.button.disabled,false);
});

test('actual clear failure retains queue and displays an error',async()=>{
 const app=appFixture({ClearAllTasksAndCaches:async()=>{throw new Error('disk denied')}});
 app.tasks.set('old',{id:'old',status:'paused'});
 await app.handlers.click();
 assert.equal(app.tasks.size,1);
 assert.equal(app.alerts.length,1);
 assert.equal(app.button.disabled,false);
});
