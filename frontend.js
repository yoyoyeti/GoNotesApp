$(
  function(){
    var signupSubmitButton = $("#signupSubmitButton");
    var loginSubmitButton = $("#loginSubmitButton");
    var username = $("#username");
    var password = $("#password");
    var message = $("#message");

    signupSubmitButton.on("click", function(){
      if(username.val().length > 0 && password.val().length > 0){
        $.ajax({
          type: 'POST',
          contentType: 'application/json; charset=utf-8',
          //url: "http://api.firstruleoffight.club/signup",
          url: "http://localhost:8080/signup",
          dataType: 'html',
          data: JSON.stringify({
            "username":username.val(),
            "password":password.val()
          }),
          success: function(result){
            switch (result) {
              case "error":
                message.html("error signing up");
                break;
              case "missing data":
                message.html("it looks like you're missing some required data");
                break;
              case "taken":
                message.html("this username is already taken, please try another");
                break;
              case "success":
                message.html("success, logging you in now");
                login();
                break;
              default:
                message.html("sorry, but it looks like something is broken");
            }
          }
        });
      }
      else {
        message.html("make sure you fill out both fields");
      }
    })

    loginSubmitButton.on("click", function () {
      login();
    })

    function login(){
      $.ajax({
        type: 'POST',
        contentType: 'application/json; charset=utf-8',
        //url: "http://api.firstruleoffight.club/login",
        url: "http://localhost:8080/login",
        dataType: 'html',
        data: JSON.stringify({
          "username":username.val(),
          "password":password.val()
        }),
        success: function(result){
          createCookie("authToken", result, 365);
          window.location.replace("http://firstruleoffight.club/notes/notes.html");
        }
      });
    }
  }
)

function createCookie(name,value,days) {
  if (days) {
    var date = new Date();
    date.setTime(date.getTime()+(days*24*60*60*1000));
    var expires = "; expires="+date.toGMTString();
  }
  else var expires = "";
  document.cookie = name+"="+value+expires+"; path=/";
}

function readCookie(name) {
  var nameEQ = name + "=";
  var ca = document.cookie.split(';');
  for(var i=0;i < ca.length;i++) {
    var c = ca[i];
    while (c.charAt(0)==' ') c = c.substring(1,c.length);
    if (c.indexOf(nameEQ) == 0) return c.substring(nameEQ.length,c.length);
  }
  return null;
}

function eraseCookie(name) {
  createCookie(name,"",-1);
}
