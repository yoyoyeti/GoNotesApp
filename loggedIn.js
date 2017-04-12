$(
  function(){

    var addNote = $("#addNote");
    var errors = $("#errors");
    var logOut = $("#logout");
    var notesDiv = $("#notesDiv");
    var textArea = $("#textArea");

    if(readCookie("authToken") == null){
      window.location.replace("http://firstruleoffight.club/notes");
    } else {
      displayNotes(readCookie("authToken"));
    }

    function displayNotes(token){
      $.ajax({
        type: 'GET',
        contentType: 'application/json; charset=utf-8',
        //url: "http://api.firstruleoffight.club/login",
        url: "http://localhost:8080/notes/" + token,
        dataType: 'html',

        success: function(result){
          if(result != "error"){
            var notes = JSON.parse(result);

            if(notes.length > 0){

              for (var i = 0; i < notes.length; i++) {
                addNoteToScreen(notes[i]);
              }
            }
          } else {
            errors.html("error displaying notes");
          }
        }
      });
    }

    function addNoteToScreen(note){
      notesDiv.append("<div><hr><span class='note'>" + note + "</span><button class='delete'>delete</button></div>");
      $(".delete").on("click", function(){
        deleteNote(this.previousElementSibling.innerHTML, this.parentNode);

      });
    }

    function deleteNote(text, nodeToRemove){
      var authToken = readCookie("authToken");
      if (authToken != null) {
        $.ajax({
          type: 'DELETE',
          contentType: 'application/json; charset=utf-8',
          //url: "http://api.firstruleoffight.club/login",
          url: "http://localhost:8080/notes",
          dataType: 'html',
          data: JSON.stringify({
            "token":authToken,
            "text":text
          }),
          success: function(result){
            if (result == "success") {
              nodeToRemove.remove();
            } else {
              errors.html("something went wrong :O");
            }
          }
        });
      } else {
        errors.html("it doesn't look like you're signed in");
      }
    }

    addNote.on("click", function(){
      var authToken = readCookie("authToken");
      var note = textArea.val();
      if(authToken != null){
        if(note.length > 0){
          $.ajax({
            type: 'POST',
            contentType: 'application/json; charset=utf-8',
            //url: "http://api.firstruleoffight.club/login",
            url: "http://localhost:8080/notes",
            dataType: 'html',
            data: JSON.stringify({
              "token":authToken,
              "text":note
            }),
            success: function(result){
              if (result == "success") {
                addNoteToScreen(note);
              } else {
                errors.html("something went wrong :O");
              }
            }
          });
        }
      } else {
        errors.html("it doesn't look like you're signed in");
      }
    })

    logOut.on("click", function(){
      eraseCookie("authToken");
      window.location.replace("http://firstruleoffight.club/notes");
    });
  }
)
