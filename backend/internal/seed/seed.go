// Package seed наполняет пустую БД демонстрационными данными,
// чтобы проверяющий сразу видел работающую ленту и мог пройти сценарий.
package seed

import (
	"log"

	"petmatch.local/backend/internal/model"
	"petmatch.local/backend/internal/repository"
	"petmatch.local/backend/internal/service"
)

// Демо-учётные записи. Они перечислены в README.md.
const (
	DemoOwnerEmail = "owner@petmatch.demo"
	DemoUserEmail  = "user@petmatch.demo"
	DemoPassword   = "demo12345"
)

type demoPet struct {
	name, typ, breed, description, photo string
	age                                  int
}

var demoPets = []demoPet{
	{"Рич", "dog", "Золотистый ретривер", "Добродушный щенок, обожает прогулки на траве и мячики. Привит, приучен к поводку.", "/assets/pets/pet1.jpg", 1},
	{"Пятнашка", "cat", "Трёхцветная кошка", "Ласковая и любопытная красавица. Стерилизована, приучена к лотку, ладит с детьми.", "/assets/pets/pet2.jpg", 2},
	{"Кучерявка", "dog", "Пудель-метис", "Кудрявая малышка с грустными глазами. Очень тянется к людям, ищет тёплый плед и заботливых хозяев.", "/assets/pets/pet3.jpg", 1},
	{"Рыжик", "cat", "Сибирская кошка", "Пушистый гордый красавец. Спокойный, независимый, но обожает, когда его чешут за ушком.", "/assets/pets/pet4.jpg", 3},
}

// Run создаёт демо-данные, только если в базе ещё нет ни одного аккаунта.
func Run(accounts *repository.AccountRepository, pets *repository.PetRepository, auth *service.AuthService, petSvc *service.PetService) error {
	n, err := accounts.Count()
	if err != nil {
		return err
	}
	if n > 0 {
		return nil
	}

	owner, err := auth.Register(DemoOwnerEmail, DemoPassword, "user", "Мария")
	if err != nil {
		return err
	}
	if err := accounts.Update(owner.Account.ID, "Мария",
		"Пристраиваю в добрые руки подобранных питомцев.", "Москва"); err != nil {
		return err
	}

	user, err := auth.Register(DemoUserEmail, DemoPassword, "user", "Анна")
	if err != nil {
		return err
	}
	if err := accounts.Update(user.Account.ID, "Анна", "Хочу подружиться с кем-нибудь пушистым.", "Москва"); err != nil {
		return err
	}

	for _, p := range demoPets {
		created, err := petSvc.CreatePet(model.Pet{
			OwnerID: owner.Account.ID, Name: p.name, Type: p.typ, Breed: p.breed,
			Age: p.age, Description: p.description, AdoptionMode: model.AdoptionOpen,
		})
		if err != nil {
			return err
		}
		if err := pets.AddPetPhoto(created.ID, p.photo); err != nil {
			return err
		}
	}
	log.Printf("seed: создан хозяин %s и пользователь %s (пароль %s)", DemoOwnerEmail, DemoUserEmail, DemoPassword)
	return nil
}
