package utils

const PageSize = 4096


const parentIDSize = 64
const prevIDSize = 64
const nextIDSize = 64
const childIDSize = 64

const numKeySize = 32
const numChildrenSize = 32
const numValSize = 32

const keySize = 64
const valSize = 64 // THIS MUST REFLECT THE SIZE OF THE V STRUCT

const internalHeaderSize = parentIDSize + numKeySize + numChildrenSize

const leafHeaderSize = parentIDSize + prevIDSize + nextIDSize + numKeySize + numValSize


const OptimalInternalOrder = (PageSize - internalHeaderSize) / (keySize + childIDSize)
const OptimalLeafOrder = (PageSize - leafHeaderSize) / (keySize + valSize)

