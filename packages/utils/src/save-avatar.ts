import fs from 'fs'

const appendAvatarFile = async (path: string, data: Buffer) => {
    try {
        await fs.promises.appendFile(path, data)
        return true
    } catch (e) {
        console.log('[SaveAvatar/appendAvatarFile]', e)
        return false
    }
}

const replaceAvatarFile = async (path: string, data: Buffer) => {
    try {
        await fs.promises.writeFile(path, data)
        return true
    } catch (e) {
        console.log('[SaveAvatar/replaceAvatarFile]', e)
        return false
    }
}

const isFileExist = async (path: string) => {
    try {
        await fs.promises.access(path, fs.constants.F_OK)
        return true
    } catch (e) {
        // console.log('[SaveAvatar/isFileExist]', e)
        return false
    }
}

const saveAvatarFile = async (path: string, data: Buffer) => {
    const isExist = await isFileExist(path)
    // console.log(path, isExist)
    if (isExist) {
        return await replaceAvatarFile(path, data)
    }
    return await appendAvatarFile(path, data)
    // await fs.promises.writeFile(uploadPath, req.file.buffer)
}

export { appendAvatarFile, replaceAvatarFile, saveAvatarFile }
