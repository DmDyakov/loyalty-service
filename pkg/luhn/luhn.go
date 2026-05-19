// Package luhn реализует проверку номеров заказов с помощью алгоритма Луна.
//
// Алгоритм Луна — это простая формула контрольной суммы,
// используемая для проверки различных идентификационных номеров,
// таких как номера кредитных карт и заказов.
//
// Подробнее: https://ru.wikipedia.org/wiki/Алгоритм_Луна
package luhn

func Valid(number string) bool {
	if len(number) == 0 {
		return false
	}

	sum := 0
	parity := len(number) % 2

	for i, r := range number {
		if r < '0' || r > '9' {
			return false
		}

		digit := int(r - '0')

		if i%2 == parity {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
	}

	return sum%10 == 0
}
