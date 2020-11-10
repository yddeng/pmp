var httpAddr = "http://127.0.0.1:9528";

function showTips(msg,t) {
    let tip = document.getElementById('tips');
    tip.style.display = "block";

    let m = document.getElementById("tips-msg");
    m.innerText=msg;
    setTimeout(function(){ tip.style.display = "none"},t)
}

function replaceLogic(logic) {
    return logic.replace(/\./g,"-")
}

function stCmd(cmd , tt, logic) {
    let req = {cmd:cmd,type:tt,logic:logic};
    let url = httpAddr+"/cmd";
    util.httpPost(url,JSON.stringify(req),function (res) {
        if (res.ok) {
            showTips("操作成功",1000)
        }else {
            showTips(res.msg,1000)
        }
    },function (e) {
        showTips("网络错误！",1000)
    })
}

function addRemCmd(cmd , tt, logic,data,success) {
    clearInterval(ticker);
    let req = {cmd:cmd,type:tt,logic:logic,data:data};
    let url = httpAddr+"/cmd";
    util.httpPost(url,JSON.stringify(req),function (res) {
        if (res.ok) {
            showTips("操作成功",1000);
            if (success){
                success();
            }
        }else {
            showTips(res.msg,2000)
        }
    },function (e) {
        showTips("网络错误！",1000)
    })
}

function start(tt,logic) {
    stCmd("start",tt,logic)
}
function stop(tt,logic) {
    stCmd("stop",tt,logic)
}
