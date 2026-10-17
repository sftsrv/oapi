package theme

import lg "charm.land/lipgloss/v2"

// color reference: https://codehs.com/uploads/7c2481e9158534231fcb3c9b6003d6b3

var ColorBlack = lg.Color("0")
var ColorWhite = lg.Color("15")
var ColorFaded = lg.Color("15")
var ColorPrimary = lg.Color("183")
var ColorSecondary = lg.Color("105")

var ColorWarn = lg.Color("220")
var ColorError = lg.Color("160")

var Heading = lg.NewStyle().Bold(true).Foreground(ColorPrimary).PaddingLeft(1).PaddingRight(1)
var Primary = lg.NewStyle().Foreground(ColorPrimary)
var Faded = lg.NewStyle().Foreground(ColorFaded)
var Secondary = lg.NewStyle().Foreground(ColorSecondary)
var Warn = lg.NewStyle().Foreground(ColorWarn)
var Alert = lg.NewStyle().Bold(true).PaddingLeft(1).PaddingRight(1).Background(ColorError)
