import{_ as h,u as A,r as u,o as C,c as r,a as e,b as k,f as w,i as O,F as b,d as D,w as E,v as N,g as o,t as l,p as I}from"./index.CbkLKmc7.js";const $={class:"receiver-view"},V={class:"header"},x={class:"transfers-table"},B=["onClick"],F={class:"transfers-table"},W=["onClick"],q={key:0,class:"dialog-overlay"},P={class:"dialog"},H={class:"form-group"},M={class:"form-group"},J={class:"dialog-actions"},Y={__name:"ReceiverView",setup(z){const g=O(),{callApi:i,error:v}=A(),R=u(null),p=u([]),d=u(!1),_=u(null),a=u({amount_received_usd:"",receiver_notes:""}),m=async()=>{const n=await i({query:`
      SELECT t.*, u.username as sender_name
      FROM transfers t
      JOIN users u ON t.id_user_do_transfer = u.id
      WHERE t.date_receive IS NULL
      ORDER BY t.date_do_transfer DESC
    `,params:[]});n.success&&(p.value=n.data)},c=(n,s)=>(n/s).toFixed(2),T=n=>{_.value=n,a.value={amount_received_usd:c(n.amount_sending_da,n.rate),receiver_notes:""},d.value=!0},y=async()=>{if(!a.value.amount_received_usd){v.value="Please fill all required fields";return}const n=await i({query:`
      UPDATE transfers 
      SET amount_received_usd = ?,
          receiver_notes = ?,
          date_receive = NOW(),
          id_user_receive_transfer = ?
      WHERE id = ? AND date_receive IS NULL
    `,params:[a.value.amount_received_usd,a.value.receiver_notes||null,R.value.id,_.value.id]});n.success?(d.value=!1,_.value=null,m(),f()):v.value=n.message},S=u([]),f=async()=>{const n=await i({query:`
      SELECT t.*, u.username as sender_name
      FROM transfers t
      JOIN users u ON t.id_user_do_transfer = u.id
      WHERE t.date_receive IS NOT NULL
      AND t.date_receive >= DATE_SUB(CURDATE(), INTERVAL 7 DAY)
      ORDER BY t.date_receive DESC
    `,params:[]});n.success&&(S.value=n.data)};C(()=>{const n=localStorage.getItem("user");n&&(R.value=JSON.parse(n),m(),f())});const L=async n=>{if(!confirm("Are you sure you want to mark this transfer as unreceived?"))return;const s=await i({query:`
      UPDATE transfers 
      SET amount_received_usd = NULL,
          receiver_notes = NULL,
          date_receive = NULL,
          id_user_receive_transfer = NULL
      WHERE id = ?
    `,params:[n.id]});s.success?(m(),f()):v.value=s.message};return(n,s)=>(o(),r("div",$,[e("div",V,[s[4]||(s[4]=e("h1",null,"All Pending Transfers",-1)),e("button",{onClick:s[0]||(s[0]=t=>w(g).push("/transfers")),class:"back-btn"},"← Return to Transfers")]),e("div",x,[e("table",null,[s[5]||(s[5]=e("thead",null,[e("tr",null,[e("th",null,"Sender"),e("th",null,"Date Sent"),e("th",null,"Sent USD"),e("th",null,"Received USD"),e("th",null,"Notes"),e("th",null,"Receiver Notes"),e("th",null,"Actions")])],-1)),e("tbody",null,[(o(!0),r(b,null,D(p.value,t=>(o(),r("tr",{key:t.id},[e("td",null,l(t.sender_name),1),e("td",null,l(new Date(t.date_do_transfer).toLocaleString()),1),e("td",null,"$"+l(c(t.amount_sending_da,t.rate)),1),e("td",null,"$"+l(t.amount_received_usd||"-"),1),e("td",null,l(t.notes||"-"),1),e("td",null,l(t.receiver_notes||"-"),1),e("td",null,[e("button",{onClick:U=>T(t),class:"btn receive-btn"}," Receive ",8,B)])]))),128))])])]),s[10]||(s[10]=e("h2",{class:"section-title"},"Recently Received Transfers (Past Week)",-1)),e("div",F,[e("table",null,[s[6]||(s[6]=e("thead",null,[e("tr",null,[e("th",null,"Sender"),e("th",null,"Date Received"),e("th",null,"Sent USD"),e("th",null,"Received USD"),e("th",null,"Notes"),e("th",null,"Receiver Notes"),e("th",null,"Actions")])],-1)),e("tbody",null,[(o(!0),r(b,null,D(S.value,t=>(o(),r("tr",{key:t.id,class:I({"amount-mismatch":c(t.amount_sending_da,t.rate)!==t.amount_received_usd})},[e("td",null,l(t.sender_name),1),e("td",null,l(new Date(t.date_receive).toLocaleString()),1),e("td",null,"$"+l(c(t.amount_sending_da,t.rate)),1),e("td",null,"$"+l(t.amount_received_usd),1),e("td",null,l(t.notes||"-"),1),e("td",null,l(t.receiver_notes||"-"),1),e("td",null,[e("button",{onClick:U=>L(t),class:"btn unreceive-btn"}," Unreceive ",8,W)])],2))),128))])])]),d.value?(o(),r("div",q,[e("div",P,[s[9]||(s[9]=e("h2",null,"Receive Transfer",-1)),e("div",H,[s[7]||(s[7]=e("label",null,"Amount Received (USD):",-1)),E(e("input",{type:"number","onUpdate:modelValue":s[1]||(s[1]=t=>a.value.amount_received_usd=t),class:"input-field",step:"0.01"},null,512),[[N,a.value.amount_received_usd]])]),e("div",M,[s[8]||(s[8]=e("label",null,"Receiver Notes:",-1)),E(e("textarea",{"onUpdate:modelValue":s[2]||(s[2]=t=>a.value.receiver_notes=t),class:"input-field",placeholder:"Add your notes here"},null,512),[[N,a.value.receiver_notes]])]),e("div",J,[e("button",{onClick:y,class:"btn receive-btn"},"Confirm"),e("button",{onClick:s[3]||(s[3]=t=>d.value=!1),class:"btn cancel-btn"},"Cancel")])])])):k("",!0)]))}},G=h(Y,[["__scopeId","data-v-949a0d73"]]);export{G as default};
//# sourceMappingURL=ReceiverView.C2Woaigj.js.map
