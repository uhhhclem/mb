'use strict';

/* Controllers */

var goods = [
	'Hides',
	'Chert',
	'Feathers',
	'Copper',
	'Mica',
	'Chalcedony',
	'Pipestone',
	'Obsidian',
	'Seashells'
];

var tribes = [
	'HoChunk',
	'Shawnee',
	'Cherokee',
	'Natchez',
	'Caddo',
	'Spanish',
];
function httpError(data, status, headers, config) {
	console.log("httpError " + status);
}

function BoardCtrl($scope, $http) {
	$scope.foo = "bar";
	$http.get("/mb/board/").success(function(data, status){
		console.log("board JSON:")
		console.log(data)
		$scope.board = new board(data);
	}).error(httpError);
	$http.get("/mb/log/").success(function(data, status) {
		$scope.log = data;
	})
}
BoardCtrl.$inject = ["$scope", '$http'];

var board = function(data){
	//TODO: deal with CaddoOrShawnee
	var modifier = null
	if (data.Card.Modifier) {
		modifier = tribes[data.Card.Modifier]
		modifier += data.Card.IsAscendant ? " + 1" : " - 1"
	}
	this.card = {
		title: data.Card.Title,
		actionPoints: data.Card.ActionPoints,
		isWhite: data.Card.IsWhite,
		resourceBonus: data.Card.ResourceBonus,
		revolt: data.Card.Revolt,
		modifier: modifier,
	};

	var lands = [[],[],[],[],[]];
	for (var i = 0; i < data.Lands.length; i++) {
		var item = data.Lands[i];
		var land = {
			name: item.Name,
			isWilderness: item.isWilderness,
			isControlled: item.isControlled
		};
		lands[item.Warpath][item.Space-1] = land;
	}
	for (var i = 0; i < data.Chiefdoms.length; i++) {
		var item = data.Chiefdoms[i];
		if (item != null) {
			var face = item.IsMounded ? item.Counter.Mounded : item.Counter.Plain;
			var chiefdom = {
				good: goods[item.Counter.Good],
				value: face.Value,
				isGreenBird: face.IsGreenBird,
			};
			lands[Math.floor(i/6)][i%6].chiefdom = chiefdom;
		}
	}
	this.lands = lands;

}