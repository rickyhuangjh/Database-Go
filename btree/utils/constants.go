package utils

const PageSize = 4096

const IDSize = 64

const parentIDSize = 64
const prevIDSize = 64
const nextIDSize = 64
const childIDSize = 64

const numKeySize = 64

const keySize = 64
const valSize = 64 // THIS MUST REFLECT THE SIZE OF THE V STRUCT

const internalHeaderSize = IDSize + parentIDSize + numKeySize

const leafHeaderSize = IDSize + parentIDSize + prevIDSize + nextIDSize + numKeySize


// optimal orders
const InternalOrder = (PageSize - internalHeaderSize) / (keySize + childIDSize)
const LeafOrder = (PageSize - leafHeaderSize) / (keySize + valSize)

