package shared

// Errors for error packets

var ERROR_0 = []byte{0, 0}
var ERROR_2 = []byte{0, 2}
var ERROR_3 = []byte{0, 3}
var ERROR_4 = []byte{0, 4}
var ERROR_5 = []byte{0, 5}
var ERROR_6 = []byte{0, 6}

// ERROR_0 is a user defined error
const ERROR_2_MESSAGE = "Access violation"
const ERROR_3_MESSAGE = "Disk full or allocation exceeded"
const ERROR_4_MESSAGE = "Illegal TFTP operation"
const ERROR_6_MESSAGE = "File already exists"

const PORT = ":8247"
