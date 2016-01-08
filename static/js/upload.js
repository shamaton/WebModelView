
var G_files="";
$(function () {
	var showFiles = function (files) {
		if (files.length==0) return false;
		//ファイル名一覧を作る
		var filelist = "";
		for(var i=0; i<files.length; i++){
			filelist += "&nbsp;&nbsp;"+files[i].name + "<br>";
		}
		//document.getElementById("mylist").innerHTML = filelist;
		$("#mylist").html(filelist);
		return false;
	};

	var uploadFiles = function (files) {
		if (files=="") return;

		document.body.style.cursor = 'wait';

		// FormData オブジェクトを用意
		var fd = new FormData();

		// ファイル情報を追加する
		for (var i = 0; i < files.length; i++) {
			fd.append("tfiles[]", files[i]);
		}

		// XHR で送信
		$.ajax({
			url: "/upload", //formではaction=に記述する名前，省略時は自分自身
			type: "POST",
			data: fd,
			mode: 'multiple',
			processData: false,
			contentType: false,
			xhr : function(){
				XHR = $.ajaxSettings.xhr();
				if(XHR.upload){
					XHR.upload.addEventListener('progress',function(e){
						progre = parseInt(e.loaded/e.total*100);
						$("#progress_bar").width(progre*5+"px");
						$("#progress_bar").html("&nbsp;"+progre+"%");
					}, false);
				}
				return XHR;
			},
			timeout: 10000,  // 単位はミリ秒
			error: function(XMLHttpRequest, textStatus, errorThrown){
				err = XMLHttpRequest.status + " : " + XMLHttpRequest.statusText + "\n" + XMLHttpRequest.responseText;
				alert(err);

				errmsg = "<font color='#ff0000' size='5'>" + XMLHttpRequest.responseText + "</font>"
				//urlが不正など通信以上の場合は 404 not found などが表示される
				//urlが正しいが返信内容が不正の場合は 200 OK が返されるが
				document.body.style.cursor = 'auto';
				$("#sendfiles").attr('disabled', false);
				$("#progress_bar").width("0px");
				$("#progress_bar").html("");
			    $("#error").html(errmsg);
			},
			beforeSend: function(xhr){
				// xhr.overrideMimeType("text/html;charset=Shift_JIS");
				//ajaxのテキスト受信がShift_JISであることを設定
				//作業が終わった時のデータの受信でSHIFT-JISが使われているため
				//utf8が使われていれば，この作業は不要
				$("#sendfiles").attr('disabled', true);
				//ボタンを無効にして二重送信防止
			}
		})
		.done(function( res ) {
		    //console.log(res);
		    $("#message").html("URLが生成されました！");
			$("#report").html("<a href='" + res + "'>" + res + "</a>");
			document.body.style.cursor = 'auto';
			$("#sendfiles").attr('disabled', false);
			$("#progress_bar").width("0px");
			$("#progress_bar").html("");
			$("#error").html("");
			$("#cancel").click();
		})
		.fail(function( res ) {
		    $("#message").html(res);
		});
	};

	// ファイル選択フォームからの入力
	$("#fileselect").bind("change", function () {
		// 選択されたファイル情報を取得
		G_files = this.files;
		//ファイル表示
		showFiles(G_files);
	});

	//送信ボタン
	$("#sendfiles").bind("click", function (e) {
		// アップロード処理
		uploadFiles(G_files);
	});

	//キャンセルボタン
	$("#cancel").bind("click", function (e) {
		//document.getElementById("mylist").innerHTML = "&nbsp;&nbsp;&nbsp;Drop File(s)";
		$("#mylist").html("&nbsp;&nbsp;&nbsp;ここにファイルをまとめてドロップ");
		G_files="";
		//document.getElementById("fileselect").value = "";
		$("#fileselect")[0].value = "";
	});

	// ドラッグドロップからの入力
	$("#target").bind("drop", function (e) {
		// ドラッグされたファイル情報を取得
		G_files = e.originalEvent.dataTransfer.files;
		//ファイル表示
		showFiles(G_files);
		// false を返してデフォルトの処理を実行しないようにする
		return false;
	})
	.bind("dragenter", function () {
		// false を返してデフォルトの処理を実行しないようにする
		return false;
	})
	.bind("dragover", function () {
		// false を返してデフォルトの処理を実行しないようにする
		return false;
	});
});