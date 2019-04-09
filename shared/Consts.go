package shared

// Errors for error packets

var Error0 = []byte{0, 0}
var Error2 = []byte{0, 2}
var Error3 = []byte{0, 3}
var Error4 = []byte{0, 4}
var Error5 = []byte{0, 5}
var Error6 = []byte{0, 6}
var Error8 = []byte{0, 8}

// ERROR_0 is a user defined error
const Error2Message = "Access violation"
const Error3Message = "Disk full or allocation exceeded"
const Error4Message = "Illegal TFTP operation"
const Error6Message = "File already exists"
const Error8Message = "Option value not supported"

const PORT = ":8247"

const MaxWindowSize = 1024
