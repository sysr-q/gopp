//gopp:define whence for
//gopp:define whither if
//gopp:define hither else
//gopp:define proclaim fmt.Println
//gopp:define proclaimeth fmt.Print
//gopp:define inquire fmt.Scanf
//gopp:define realm package
//gopp:define desire import
//gopp:define activity func
//gopp:define tis =
//gopp:define be :=
//gopp:define ye {
//gopp:define desist }
//gopp:define meaning var
//gopp:define platitude bool
//gopp:define sooth true
//gopp:define falsehood false
//gopp:define nay !

realm main

desire (
	"fmt"
)

activity main() ye
	meaning (
		done platitude
		chooser int
		name, item string
	)

	whence nay done ye
		done tis sooth
		proclaim("Hi! Chooseth the number of the application you wish to run.")
		proclaim("1. What is your name?")
		proclaim("2. Pluralizer.")
		proclaim("3. Shakespeare.")
		proclaimeth("> ")
		inquire("%d", &chooser)

		whither chooser == 1 ye
			proclaim("What is your name?")
			proclaimeth("> ")
			inquire("%s", &name)
			proclaim("Hello,", name + "!")
		desist hither whither chooser == 2 ye
			proclaim("What singular item would you like more of?")
			proclaimeth("> ")
			inquire("%s", &item)
			proclaim("Then buy more", item + "s.")
		desist hither whither chooser == 3 ye
			proclaim("\"To be or not to be,\n that is the question.\"")
		desist hither ye
			proclaim("That is not a valid option.\n")
			done tis falsehood
		desist
	desist
desist
