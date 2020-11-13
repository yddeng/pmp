var util = {};
util.httpGet = function(url,success,error){
    $.ajax({
        url:url,
        type: "get",
        async: true,
        success: success,
        error: error
    });
};

util.httpPost = function(url,data,success,error){
    $.ajax({
        url:url,
        type: "post",
        async: true,
        dataType: "json",
        data:data,
        success: success,
        error: error
    });
};

util.httpFormData = function(url,data,success,error){
    $.ajax({
        url:url,
        type: "post",
        data:data,
        contentType: false,
        processData: false,
        success: success,
        error: error
    });
};

//弹出一个询问框，有确定和取消按钮
util.firm = function(msg,firmFunc) {
    //利用对话框返回的值 （true 或者 false）
    if (confirm(msg) ){
        firmFunc()
    }
};

// 字符串格式化
util.format = function(src){
    if (arguments.length == 0) return null;
    let args = Array.prototype.slice.call(arguments, 1);
    return src.replace(/\{(\d+)\}/g, function(m, i){
        return args[i];
    });
};

// url
util.getUrlParam =  function(name) {
    let reg = new RegExp('(^|&)'+ name + '=([^&]*)(&|$)');
    let result= window.location.search.substr(1).match(reg);
    return result?decodeURIComponent(result[2]):null;
};

// 设置cookie的函数  （名字，值，过期时间（天））
util.setCookie = function (cname, cvalue, exdays) {
    let d = new Date();
    d.setTime(d.getTime() + (exdays * 24 * 60 * 60 * 1000));
    let expires = "expires=" + d.toUTCString();
    document.cookie = cname + "=" + cvalue + "; " + expires;
};

//获取cookie
//取cookie的函数(名字) 取出来的都是字符串类型 子目录可以用根目录的cookie，根目录取不到子目录的 大小4k左右
util.getCookie = function(cname) {
    let name = cname + "=";
    let ca = document.cookie.split(';');
    for(let i=0; i<ca.length; i++)
    {
        let c = ca[i].trim();
        if (c.indexOf(name)===0) return c.substring(name.length,c.length);
    }
    return "";
};

util.percent = function (v) {
    let n = parseFloat(v);
    return util.format("{0}%", n.toFixed(2))
};

util.str2Int = function (v) {
    return parseInt(v)
};

function showTips(msg,t) {
    let tip = document.getElementById('tips');
    tip.style.display = "block";

    let m = document.getElementById("tips-msg");
    m.innerText=msg;
    setTimeout(function(){ tip.style.display = "none"},t)
}