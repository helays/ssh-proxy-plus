<script lang="ts">
const getRandomChar=(chars:string):string=>{
    return chars.charAt(Math.floor(Math.random() * chars.length));
}

export const GeneratePassword=(remainingLength:number):string=>{
    const lowerCaseLetters = 'abcdefghijklmnopqrstuvwxyz';
    const upperCaseLetters = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ';
    const numbers = '0123456789';
    const specialChars = '()`~!@#$%^&*-_+=|{}[]:;\'<>,.?';

    // 确保密码包含所有必需字符类型
    let password = [
        getRandomChar(lowerCaseLetters),
        getRandomChar(upperCaseLetters),
        getRandomChar(numbers),
        getRandomChar(specialChars)
    ];

    // 填充剩余长度，保证总长度在8到30之间
    // const remainingLength = Math.floor(Math.random() * (30 - 4)) + 4; // 随机决定额外添加的字符数，确保总长度在8到30之间
    const allChars = lowerCaseLetters + upperCaseLetters + numbers + specialChars;

    for (let i = 0; i < (remainingLength-4); i++) {
        password.push(getRandomChar(allChars));
    }

    // 打乱密码中的字符顺序以增加随机性
    password = password.sort(() => 0.5 - Math.random());

    return password.join('');
}

export const ValidatePassword=(password:string):boolean=>{
    const hasLower = /[a-z]/.test(password);
    const hasUpper = /[A-Z]/.test(password);
    const hasNumber = /\d/.test(password);
    const hasSpecial = /[()`~!@#$%^&*-_+=|{}[\]:;'<>,.?/]/.test(password);
    const isValidLength = password.length >= 8 && password.length <= 30;

    return hasLower && hasUpper && hasNumber && hasSpecial && isValidLength;
}

export default {
}
</script>