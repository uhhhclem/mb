'use strict';

// Declare app level module which depends on filters, and services
angular.module('mb', ['mb.filters', 'mb.services', 'mb.directives']).
  config(['$routeProvider', function($routeProvider) {
  	$routeProvider.when('/board', {templateUrl: 'partials/board.html', controller: BoardCtrl});
    $routeProvider.otherwise({redirectTo: '/board'});
  }]);
