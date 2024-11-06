package main

import (	
)

type Node struct {
	Id            string  `json:"id"`            //Unique identifier of the node. This ID is referenced by edge in its source and target field.
	Title         string  `json:"title"`         //Name of the node visible in just under the node.
	MainStat      float32 `json:"mainStat"`      //First stat shown inside the node itself
	SecondaryStat float32 `json:"secondaryStat"` //Same as mainStat, but shown under it inside the node
	Arc__Failed   float32 `json:"arc__failed"`   //to create the color circle around the node. All values in these fields should add up to 1.
	Arc__Passed   float32 `json:"arc__passed"`   //
	Detail__Role  string  `json:"detail__role"`  //shown in the header of context menu when clicked on the node
	Color         string  `json:"color"`         //Can be used to specify a single color instead of using the arc__ fields to specify color sections
	Icon          string  `json:"icon"`          //
	NodeRadius    int     `json:"nodeRadius"`    //Radius value in pixels. Used to manage node size.
	Highlighted   bool    `json:"highlighted"`   //Sets whether the node should be highlighted.
}

type Edge struct {
	Id            string  `json:"id"`            //Unique identifier of the edge.
	Source        string  `json:"source"`        //Id of the source node.
	Target        string  `json:"target"`        //Id of the target.
	MainStat      float32 `json:"mainStat"`      //First stat shown in the overlay when hovering over the edge.
	SecondaryStar float32 `json:"secondarystat"` //Same as mainStat, but shown right under it.
	Detail__Info  string  `json:"detail__info"`  //will be shown in the header of context menu when clicked on the edge
	Thickness     float32 `json:"thickness"`     //The thickness of the edge. Default: 1
	Highlighted   bool    `json:"highlighted"`   //boolean	Sets whether the edge should be highlighted.
	Color         string  `json:"color"`         //string	Sets the default color of the edge. It can be an acceptable HTML color string. Default: #999
}